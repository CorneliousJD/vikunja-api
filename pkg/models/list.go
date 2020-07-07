// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/metrics"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web"
	"strings"
	"time"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// List represents a list of tasks
type List struct {
	// The unique, numeric id of this list.
	ID int64 `xorm:"int(11) autoincr not null unique pk" json:"id" param:"list"`
	// The title of the list. You'll see this in the namespace overview.
	Title string `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(1|250)" minLength:"3" maxLength:"250"`
	// The description of the list.
	Description string `xorm:"longtext null" json:"description"`
	// The unique list short identifier. Used to build task identifiers.
	Identifier string `xorm:"varchar(10) null" json:"identifier" valid:"runelength(0|10)" minLength:"0" maxLength:"10"`
	// The hex color of this list
	HexColor string `xorm:"varchar(6) null" json:"hex_color" valid:"runelength(0|6)" maxLength:"6"`

	OwnerID     int64 `xorm:"int(11) INDEX not null" json:"-"`
	NamespaceID int64 `xorm:"int(11) INDEX not null" json:"namespace_id" param:"namespace"`

	// The user who created this list.
	Owner *user.User `xorm:"-" json:"owner" valid:"-"`
	// An array of tasks which belong to the list.
	// Deprecated: you should use the dedicated task list endpoint because it has support for pagination and filtering
	Tasks []*Task `xorm:"-" json:"-"`

	// Whether or not a list is archived.
	IsArchived bool `xorm:"not null default false" json:"is_archived" query:"is_archived"`

	// The id of the file this list has set as background
	BackgroundFileID int64 `xorm:"null" json:"-"`
	// Holds extra information about the background set since some background providers require attribution or similar. If not null, the background can be accessed at /lists/{listID}/background
	BackgroundInformation interface{} `xorm:"-" json:"background_information"`

	// A timestamp when this list was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this list was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// ListBackgroundType holds a list background type
type ListBackgroundType struct {
	Type string
}

// ListBackgroundUpload represents the list upload background type
const ListBackgroundUpload string = "upload"

// GetListsByNamespaceID gets all lists in a namespace
func GetListsByNamespaceID(nID int64, doer *user.User) (lists []*List, err error) {
	if nID == -1 {
		err = x.Select("l.*").
			Table("list").
			Join("LEFT", []string{"team_list", "tl"}, "l.id = tl.list_id").
			Join("LEFT", []string{"team_members", "tm"}, "tm.team_id = tl.team_id").
			Join("LEFT", []string{"users_list", "ul"}, "ul.list_id = l.id").
			Join("LEFT", []string{"namespaces", "n"}, "l.namespace_id = n.id").
			Where("tm.user_id = ?", doer.ID).
			Where("l.is_archived = false").
			Where("n.is_archived = false").
			Or("ul.user_id = ?", doer.ID).
			GroupBy("l.id").
			Find(&lists)
	} else {
		err = x.Select("l.*").
			Alias("l").
			Join("LEFT", []string{"namespaces", "n"}, "l.namespace_id = n.id").
			Where("l.is_archived = false").
			Where("n.is_archived = false").
			Where("namespace_id = ?", nID).
			Find(&lists)
	}
	if err != nil {
		return nil, err
	}

	// get more list details
	err = AddListDetails(lists)
	return lists, err
}

// ReadAll gets all lists a user has access to
// @Summary Get all lists a user has access to
// @Description Returns all lists a user has access to.
// @tags list
// @Accept json
// @Produce json
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search lists by title."
// @Param is_archived query bool false "If true, also returns all archived lists."
// @Security JWTKeyAuth
// @Success 200 {array} models.List "The lists"
// @Failure 403 {object} web.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists [get]
func (l *List) ReadAll(a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {
	// Check if we're dealing with a share auth
	shareAuth, ok := a.(*LinkSharing)
	if ok {
		list := &List{ID: shareAuth.ListID}
		err := list.GetSimpleByID()
		if err != nil {
			return nil, 0, 0, err
		}
		lists := []*List{list}
		err = AddListDetails(lists)
		return lists, 0, 0, err
	}

	lists, resultCount, totalItems, err := getRawListsForUser(&listOptions{
		search:     search,
		user:       &user.User{ID: a.GetID()},
		page:       page,
		perPage:    perPage,
		isArchived: l.IsArchived,
	})
	if err != nil {
		return nil, 0, 0, err
	}

	// Add more list details
	err = AddListDetails(lists)
	return lists, resultCount, totalItems, err
}

// ReadOne gets one list by its ID
// @Summary Gets one list
// @Description Returns a list by its ID.
// @tags list
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "List ID"
// @Success 200 {object} models.List "The list"
// @Failure 403 {object} web.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id} [get]
func (l *List) ReadOne() (err error) {
	// Get list owner
	l.Owner, err = user.GetUserByID(l.OwnerID)
	if err != nil {
		return err
	}
	// Check if the namespace is archived and set the namespace to archived if it is not already archived individually.
	if !l.IsArchived {
		err = l.CheckIsArchived()
		if err != nil {
			if !IsErrNamespaceIsArchived(err) && !IsErrListIsArchived(err) {
				return
			}
			l.IsArchived = true
		}
	}

	// Get any background information if there is one set
	if l.BackgroundFileID != 0 {
		// Unsplash image
		l.BackgroundInformation, err = GetUnsplashPhotoByFileID(l.BackgroundFileID)
		if err != nil && !files.IsErrFileIsNotUnsplashFile(err) {
			return
		}

		if err != nil && files.IsErrFileIsNotUnsplashFile(err) {
			l.BackgroundInformation = &ListBackgroundType{Type: ListBackgroundUpload}
		}
	}

	return nil
}

// GetSimpleByID gets a list with only the basic items, aka no tasks or user objects. Returns an error if the list does not exist.
func (l *List) GetSimpleByID() (err error) {
	if l.ID < 1 {
		return ErrListDoesNotExist{ID: l.ID}
	}

	// We need to re-init our list object, because otherwise xorm creates a "where for every item in that list object,
	// leading to not finding anything if the id is good, but for example the title is different.
	id := l.ID
	*l = List{}
	exists, err := x.Where("id = ?", id).Get(l)
	if err != nil {
		return
	}

	if !exists {
		return ErrListDoesNotExist{ID: l.ID}
	}

	return
}

// GetListSimplByTaskID gets a list by a task id
func GetListSimplByTaskID(taskID int64) (l *List, err error) {
	// We need to re-init our list object, because otherwise xorm creates a "where for every item in that list object,
	// leading to not finding anything if the id is good, but for example the title is different.
	var list List
	exists, err := x.
		Select("list.*").
		Table(List{}).
		Join("INNER", "tasks", "list.id = tasks.list_id").
		Where("tasks.id = ?", taskID).
		Get(&list)
	if err != nil {
		return
	}

	if !exists {
		return &List{}, ErrListDoesNotExist{ID: l.ID}
	}

	return &list, nil
}

type listOptions struct {
	search     string
	user       *user.User
	page       int
	perPage    int
	isArchived bool
}

// Gets the lists only, without any tasks or so
func getRawListsForUser(opts *listOptions) (lists []*List, resultCount int, totalItems int64, err error) {
	fullUser, err := user.GetUserByID(opts.user.ID)
	if err != nil {
		return nil, 0, 0, err
	}

	// Adding a 1=1 condition by default here because xorm always needs a condition and cannot handle nil conditions
	var isArchivedCond builder.Cond = builder.Eq{"1": 1}
	if !opts.isArchived {
		isArchivedCond = builder.And(
			builder.Eq{"l.is_archived": false},
			builder.Eq{"n.is_archived": false},
		)
	}

	limit, start := getLimitFromPageIndex(opts.page, opts.perPage)

	// Gets all Lists where the user is either owner or in a team which has access to the list
	// Or in a team which has namespace read access
	query := x.Select("l.*").
		Table("list").
		Alias("l").
		Join("INNER", []string{"namespaces", "n"}, "l.namespace_id = n.id").
		Join("LEFT", []string{"team_namespaces", "tn"}, "tn.namespace_id = n.id").
		Join("LEFT", []string{"team_members", "tm"}, "tm.team_id = tn.team_id").
		Join("LEFT", []string{"team_list", "tl"}, "l.id = tl.list_id").
		Join("LEFT", []string{"team_members", "tm2"}, "tm2.team_id = tl.team_id").
		Join("LEFT", []string{"users_list", "ul"}, "ul.list_id = l.id").
		Join("LEFT", []string{"users_namespace", "un"}, "un.namespace_id = l.namespace_id").
		Where(builder.Or(
			builder.Eq{"tm.user_id": fullUser.ID},
			builder.Eq{"tm2.user_id": fullUser.ID},
			builder.Eq{"ul.user_id": fullUser.ID},
			builder.Eq{"un.user_id": fullUser.ID},
			builder.Eq{"l.owner_id": fullUser.ID},
		)).
		GroupBy("l.id").
		Where("l.title LIKE ?", "%"+opts.search+"%").
		Where(isArchivedCond)
	if limit > 0 {
		query = query.Limit(limit, start)
	}
	err = query.Find(&lists)
	if err != nil {
		return nil, 0, 0, err
	}

	totalItems, err = x.
		Table("list").
		Alias("l").
		Join("INNER", []string{"namespaces", "n"}, "l.namespace_id = n.id").
		Join("LEFT", []string{"team_namespaces", "tn"}, "tn.namespace_id = n.id").
		Join("LEFT", []string{"team_members", "tm"}, "tm.team_id = tn.team_id").
		Join("LEFT", []string{"team_list", "tl"}, "l.id = tl.list_id").
		Join("LEFT", []string{"team_members", "tm2"}, "tm2.team_id = tl.team_id").
		Join("LEFT", []string{"users_list", "ul"}, "ul.list_id = l.id").
		Join("LEFT", []string{"users_namespace", "un"}, "un.namespace_id = l.namespace_id").
		Where(builder.Or(
			builder.Eq{"tm.user_id": fullUser.ID},
			builder.Eq{"tm2.user_id": fullUser.ID},
			builder.Eq{"ul.user_id": fullUser.ID},
			builder.Eq{"un.user_id": fullUser.ID},
			builder.Eq{"l.owner_id": fullUser.ID},
		)).
		GroupBy("l.id").
		Where("l.title LIKE ?", "%"+opts.search+"%").
		Where(isArchivedCond).
		Count(&List{})
	return lists, len(lists), totalItems, err
}

// AddListDetails adds owner user objects and list tasks to all lists in the slice
func AddListDetails(lists []*List) (err error) {
	var ownerIDs []int64
	for _, l := range lists {
		ownerIDs = append(ownerIDs, l.OwnerID)
	}

	// Get all list owners
	owners := map[int64]*user.User{}
	err = x.In("id", ownerIDs).Find(&owners)
	if err != nil {
		return
	}

	var fileIDs []int64
	for _, l := range lists {
		if o, exists := owners[l.OwnerID]; exists {
			l.Owner = o
		}
		if l.BackgroundFileID != 0 {
			l.BackgroundInformation = &ListBackgroundType{Type: ListBackgroundUpload}
		}
		fileIDs = append(fileIDs, l.BackgroundFileID)
	}

	// Unsplash background file info
	us := []*UnsplashPhoto{}
	err = x.In("file_id", fileIDs).Find(&us)
	if err != nil {
		return
	}
	unsplashPhotos := make(map[int64]*UnsplashPhoto, len(us))
	for _, u := range us {
		unsplashPhotos[u.FileID] = u
	}

	// Build it all into the lists slice
	for _, l := range lists {
		l.BackgroundInformation = unsplashPhotos[l.BackgroundFileID]
	}

	return
}

// NamespaceList is a meta type to be able  to join a list with its namespace
type NamespaceList struct {
	List      List      `xorm:"extends"`
	Namespace Namespace `xorm:"extends"`
}

// CheckIsArchived returns an ErrListIsArchived or ErrNamespaceIsArchived if the list or its namespace is archived.
func (l *List) CheckIsArchived() (err error) {
	// When creating a new list, we check if the namespace is archived
	if l.ID == 0 {
		n := &Namespace{ID: l.NamespaceID}
		return n.CheckIsArchived()
	}

	nl := &NamespaceList{}
	exists, err := x.
		Table("list").
		Join("LEFT", "namespaces", "list.namespace_id = namespaces.id").
		Where("list.id = ? AND (list.is_archived = true OR namespaces.is_archived = true)", l.ID).
		Get(nl)
	if err != nil {
		return
	}
	if exists && nl.List.ID != 0 && nl.List.IsArchived {
		return ErrListIsArchived{ListID: l.ID}
	}
	if exists && nl.Namespace.ID != 0 && nl.Namespace.IsArchived {
		return ErrNamespaceIsArchived{NamespaceID: nl.Namespace.ID}
	}
	return nil
}

// GenerateListIdentifier generates a unique random list identifier based on the list title.
// If it is not able to find one, the list identifier will be empty.
func GenerateListIdentifier(l *List, sess *xorm.Engine) (err error) {

	// The general idea here is to take the title and slice it into pieces, until we found a unique piece.

	var exists = true
	titleSlug := []rune(strings.Replace(strings.ToUpper(l.Title), " ", "", -1))

	// We can save at most 10 characters in the db, so we need to ensure it has at most 10 characters
	if len(titleSlug) > 10 {
		titleSlug = titleSlug[0:9]
	}

	var i = 0

	for exists {

		// Prevent endless looping
		if i == len(titleSlug) {
			break
		}

		// Take a random part of the title slug, starting at the beginning
		l.Identifier = string(titleSlug[i:])
		exists, err = sess.
			Where("identifier = ?", l.Identifier).
			And("id != ?", l.ID).
			Exist(&List{})
		if err != nil {
			return
		}
		i++
	}
	return nil
}

// CreateOrUpdateList updates a list or creates it if it doesn't exist
func CreateOrUpdateList(list *List) (err error) {

	// Check if the namespace exists
	if list.NamespaceID != 0 {
		_, err = GetNamespaceByID(list.NamespaceID)
		if err != nil {
			return err
		}
	}

	// Check if the identifier is unique and not empty
	if list.Identifier != "" {
		exists, err := x.
			Where("identifier = ?", list.Identifier).
			And("id != ?", list.ID).
			Exist(&List{})
		if err != nil {
			return err
		}
		if exists {
			return ErrListIdentifierIsNotUnique{Identifier: list.Identifier}
		}
	}

	// Generate a random list identifier base on the list title
	if list.ID == 0 && list.Identifier == "" {
		err = GenerateListIdentifier(list, x)
		if err != nil {
			return
		}
	}

	if list.ID == 0 {
		_, err = x.Insert(list)
		metrics.UpdateCount(1, metrics.ListCountKey)
	} else {
		// We need to specify the cols we want to update here to be able to un-archive lists
		colsToUpdate := []string{
			"title",
			"is_archived",
			"identifier",
			"hex_color",
		}
		if list.Description != "" {
			colsToUpdate = append(colsToUpdate, "description")
		}

		_, err = x.
			ID(list.ID).
			Cols(colsToUpdate...).
			Update(list)
	}

	if err != nil {
		return
	}

	err = list.GetSimpleByID()
	if err != nil {
		return
	}

	err = list.ReadOne()
	return

}

// Update implements the update method of CRUDable
// @Summary Updates a list
// @Description Updates a list. This does not include adding a task (see below).
// @tags list
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "List ID"
// @Param list body models.List true "The list with updated values you want to update."
// @Success 200 {object} models.List "The updated list."
// @Failure 400 {object} web.HTTPError "Invalid list object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id} [post]
func (l *List) Update() (err error) {
	return CreateOrUpdateList(l)
}

func updateListLastUpdated(list *List) error {
	_, err := x.ID(list.ID).Cols("updated").Update(list)
	return err
}

func updateListByTaskID(taskID int64) (err error) {
	// need to get the task to update the list last updated timestamp
	task, err := GetTaskByIDSimple(taskID)
	if err != nil {
		return err
	}

	return updateListLastUpdated(&List{ID: task.ListID})
}

// Create implements the create method of CRUDable
// @Summary Creates a new list
// @Description Creates a new list in a given namespace. The user needs write-access to the namespace.
// @tags list
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param namespaceID path int true "Namespace ID"
// @Param list body models.List true "The list you want to create."
// @Success 200 {object} models.List "The created list."
// @Failure 400 {object} web.HTTPError "Invalid list object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{namespaceID}/lists [put]
func (l *List) Create(a web.Auth) (err error) {
	err = l.CheckIsArchived()
	if err != nil {
		return err
	}

	doer, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}

	l.OwnerID = doer.ID
	l.Owner = doer
	l.ID = 0 // Otherwise only the first time a new list would be created

	err = CreateOrUpdateList(l)
	if err != nil {
		return
	}

	// Create a new first bucket for this list
	b := &Bucket{
		ListID: l.ID,
		Title:  "New Bucket",
	}
	return b.Create(a)
}

// Delete implements the delete method of CRUDable
// @Summary Deletes a list
// @Description Delets a list
// @tags list
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "List ID"
// @Success 200 {object} models.Message "The list was successfully deleted."
// @Failure 400 {object} web.HTTPError "Invalid list object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id} [delete]
func (l *List) Delete() (err error) {

	// Delete the list
	_, err = x.ID(l.ID).Delete(&List{})
	if err != nil {
		return
	}
	metrics.UpdateCount(-1, metrics.ListCountKey)

	// Delete all todotasks on that list
	_, err = x.Where("list_id = ?", l.ID).Delete(&Task{})
	return
}

// SetListBackground sets a background file as list background in the db
func SetListBackground(listID int64, background *files.File) (err error) {
	l := &List{
		ID:               listID,
		BackgroundFileID: background.ID,
	}
	_, err = x.
		Where("id = ?", l.ID).
		Cols("background_file_id").
		Update(l)
	return
}
