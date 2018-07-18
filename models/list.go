package models

// List represents a list of items
type List struct {
	ID          int64  `xorm:"int(11) autoincr not null unique pk" json:"id" param:"listid"`
	Title       string `xorm:"varchar(250)" json:"title"`
	Description string `xorm:"varchar(1000)" json:"description"`
	OwnerID     int64  `xorm:"int(11)" json:"-"`
	NamespaceID int64  `xorm:"int(11)" json:"-" param:"nid"`

	Owner User        `xorm:"-" json:"owner"`
	Items []*ListItem `xorm:"-" json:"items"`

	Created int64 `xorm:"created" json:"created"`
	Updated int64 `xorm:"updated" json:"updated"`

	CRUDable `xorm:"-" json:"-"`
	Rights   `xorm:"-" json:"-"`
}

// AfterLoad loads the owner and list items
func (l *List) AfterLoad() {

	// Get the owner
	l.Owner, _, _ = GetUserByID(l.OwnerID)

	// Get the list items
	l.Items, _ = GetItemsByListID(l.ID)
}

// GetListByID returns a list by its ID
func GetListByID(id int64) (list List, err error) {
	exists, err := x.ID(id).Get(&list)
	if err != nil {
		return list, err
	}

	if !exists {
		return list, ErrListDoesNotExist{ID: id}
	}

	return list, nil
}

// GetListsByNamespaceID gets all lists in a namespace
func GetListsByNamespaceID(nID int64) (lists []*List, err error) {
	err = x.Where("namespace_id = ?", nID).Find(&lists)
	return lists, err
}

// ReadAll gets all lists a user has access to
func (l *List) ReadAll(user *User) (interface{}, error) {
	lists := []List{}
	fullUser, _, err := GetUserByID(user.ID)
	if err != nil {
		return lists, err
	}

	// Gets all Lists where the user is either owner or in a team which has access to the list
	// Or in a team which has namespace read access
	err = x.Select("l.*").
		Table("list").
		Alias("l").
		Join("INNER", []string{"namespaces", "n"}, "l.namespace_id = n.id").
		Join("LEFT", []string{"team_namespaces", "tn"}, "tn.namespace_id = n.id").
		Join("LEFT", []string{"team_members", "tm"}, "tm.team_id = tn.team_id").
		Join("LEFT", []string{"team_list", "tl"}, "l.id = tl.list_id").
		Join("LEFT", []string{"team_members", "tm2"}, "tm2.team_id = tl.team_id").
		Where("tm.user_id = ?", fullUser.ID).
		Or("tm2.user_id = ?", fullUser.ID).
		Or("l.owner_id = ?", fullUser.ID).
		Find(&lists)

	return lists, err
}

// ReadOne gets one list by its ID
func (l *List) ReadOne(id int64) (err error) {
	*l, err = GetListByID(id)
	return
}