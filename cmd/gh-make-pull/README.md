## Github Make Pull
Opens a pull request on GitHub against "origin/master" for the branch you are on.

If you execute it with no arguments, it expects to receive input on stdin (pipe) the pull request title and description.
The text up to the first blank line on stdin is treated as the pull request title, and the rest is used as pull request description in Markdown format.

If executed with an argument, it uses that as the title, and any stdin is just the decription.

This command will abort operation if it detects that any of the following are true:
+ The current branch is master.
+ the current topic branch has local commits that are not yet pushed to its upstream branch on the remote.
+ the current repository has uncommitted files

Assuming you are in a git repo with "origin" remote pointing to a github repo, you've committed your changes,
 and you are on a branch that has been pushed:
```
$ gh-make-pull <title>
```

### How does it work

GitHub's API requires authentication, the simplest way is to use a Personal Access Token, and setting the environment's GITHUB_TOKEN to this value.

```
export GITHUB_PULL_REQUEST_AUTH_TOKEN=aabbcc...ddeeff
```
make_pull then takes the GitHub's user's login as the first and only argument, and it's required.

```
make_pull mynameisawesome
```

## Personal Access tokens

Check them out here https://github.com/settings/tokens


### OAuth

From [this gist](https://gist.github.com/btoone/2288960)

The first thing to know is that your API Token (found in https://github.com/settings/admin) is not the same token used by OAuth. They are different tokens and you will need to generate an OAuth token to be authorized.

Follow the API's instructions at http://developer.github.com/v3/oauth/ under the sections "Non-Web Application Flow" and "Create a new authorization" to become authorized.

Note: Use Basic Auth once to create an OAuth2 token http://developer.github.com/v3/oauth/#oauth-authorizations-api

    curl https://api.github.com/authorizations \
    --user "caspyin" \
    --data '{"scopes":["gist"],"note":"Demo"}'

This will prompt you for your GitHub password and return your OAuth token in the response. It will also create a new Authorized application in your account settings https://github.com/settings/applications

Now that you have the OAuth token there are two ways to use the token to make requests that require authentication (replace "OAUTH-TOKEN" with your actual token)

    curl https://api.github.com/gists/starred?access_token=OAUTH-TOKEN
    curl -H "Authorization: token OAUTH-TOKEN" https://api.github.com/gists/starred

List the authorizations you already have

    curl --user "caspyin" https://api.github.com/authorizations

## Make a pull request

 https://developer.github.com/v3/pulls/#create-a-pull-request
 
 `POST /repos/:owner/:repo/pulls`
 ```
 {
   "title": "Amazing new feature",
   "body": "Please pull this in!",
   "head": "octocat:new-feature",
   "base": "master"
 }
 ```
