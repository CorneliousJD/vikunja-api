<img src="https://vikunja.io/images/vikunja-logo.svg" alt="" style="display: block;width: 50%;margin: 0 auto;" width="50%"/>

[![Build Status](https://drone1.kolaente.de/api/badges/vikunja/api/status.svg)](https://drone1.kolaente.de/vikunja/api)
[![License: GPL v3](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](LICENSE)
[![Download](https://img.shields.io/badge/download-v0.15.1-brightgreen.svg)](https://dl.vikunja.io)
[![Docker Pulls](https://img.shields.io/docker/pulls/vikunja/api.svg)](https://hub.docker.com/r/vikunja/api/)
[![Swagger Docs](https://img.shields.io/badge/swagger-docs-brightgreen.svg)](https://try.vikunja.io/api/v1/docs)
[![Go Report Card](https://goreportcard.com/badge/git.kolaente.de/vikunja/api)](https://goreportcard.com/report/git.kolaente.de/vikunja/api)

# Vikunja API

> The Todo-app to organize your life.

# Table of contents

* [Features](#features)
* [Docs](#docs)
* [Roadmap](#roadmap)
* [Contributing](#contributing)
* [License](#license)

## Features

* Create TODO lists with tasks
  * Reminder for tasks
* Namespaces: A "group" which bundels multiple lists
* Share lists and namespaces with teams and users with granular permissions
* Plenty of details for tasks

See [the features page](https://vikunja.io/en/features/) on our website for a more exaustive list or 
try it on [try.vikunja.io](https://try.vikunja.io)!

## Docs

* [Installing](https://vikunja.io/docs/installing/)
* [Build from source](https://vikunja.io/docs/build-from-sources/)
* [Development setup](https://vikunja.io/docs/development/)
* [Magefile](https://vikunja.io/docs/mage/)
* [Testing](https://vikunja.io/docs/testing/)

All docs can be found on [the vikunja home page](https://vikunja.io/docs/).

### Roadmap

> I know, it's still a long way to go. I'm currently working on a lot of "basic" features, the exiting things will come later. Don't worry, they'll come.

* [x] Prioritize tasks
* [x] Subtasks
* [x] Repeating tasks
* [x] Get tasks via caldav
* [x] Get all your tasks for an interval (day/month/period)
* [x] Labels for tasks
* [x] Assign users to tasks
* [x] Attachments on tasks
* [x] More sharing features
  * [x] Share with individual users
  * [x] Share via a world-readable link with or without password, like Nextcloud
* [x] Disable registration, making an instance "invite-only" 
* [ ] SSE to notify multiple clients of updates when something was changed
* [ ] "Smart Lists" - Create lists based on filters
* [ ] IMAP-Integration - Send an email to Vikunja to create a new task
* [ ] Webhooks - Trigger other events when an action is done (like completing a task)
* [ ] Performace statistics - Get an overview and beautiful charts about what you got done this month
* [ ] Activity feeds - Get a quick overview about who did what
* [ ] Bulk-edit multiple tasks at once
* [ ] Team-efforts - Requiring a task to be marked as done by multiple members until it's done
* [ ] Global limits for namespaces/lists/tasks

See [our roadmap](https://my.vikunja.cloud/share/QFyzYEmEYfSyQfTOmIRSwLUpkFjboaBqQCnaPmWd/auth) (hosted on Vikunja!) for even more!

* [ ] [Mobile apps](https://code.vikunja.io/app) (seperate repo) *In Progress*
* [ ] [Webapp](https://code.vikunja.io/frontend) (seperate repo) *In Progress*

## Contributing

Fork -> Push -> Pull-Request. Also see the [dev docs](https://vikunja.io/docs/development/) for more infos.

## License

This project is licensed under the GPLv3 License. See the [LICENSE](LICENSE) file for the full license text.
