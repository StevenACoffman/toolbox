### Get all boards
curl 'https://khanacademy.atlassian.net/rest/agile/1.0/board/' --user $(whoami)@khanacademy.org:${JIRA_API_TOKEN} | jq .

### Get board
curl 'https://khanacademy.atlassian.net/rest/agile/1.0/board/233' --user $(whoami)@khanacademy.org:${JIRA_API_TOKEN} | jq .

### Get Sprints

/rest/agile/1.0/board/{boardId}/sprint

###  Get epics
GET /rest/agile/1.0/board/{boardId}/epic


### Get all isssues for an epic:
GET /rest/agile/1.0/epic/{epicIdOrKey}/issue

Get all issues for a sprint:
curl 'https://khanacademy.atlassian.net/rest/agile/1.0/sprint/1032/issue' --user $(whoami)@khanacademy.org:${JIRA_API_TOKEN} | jq .

