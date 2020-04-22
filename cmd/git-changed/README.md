### git-changed
Equivalent to `git diff --name-only`

This is pretty handy in a monorepo, as are:
+ [GTA](https://github.com/jphines/gta)
+ [Affected](https://github.com/jharlap/affected)
+ [sowhatsnew](https://github.com/manifoldco/sowhatsnew/)
+ [mono-meta](https://github.com/davidae/mono-meta)
+ [knega](https://github.com/kristofferlind/knega)
+ [mona](https://github.com/davidsbond/mona)
+ [detect-changed-services](https://github.com/devops-recipes/app-mono/blob/master/detect-changed-services.sh)
+ [golang-monorepo](https://github.com/flowerinthenight/golang-monorepo)
+ [baur](https://github.com/simplesurance/baur)
+ [monorepo-operator](https://github.com/SimonBaeumer/monorepo-operator)
+ [mbt](https://github.com/mbtproject/mbt)
+ [drone-tree-config](https://github.com/bitsbeats/drone-tree-config) (github pr based dir diff)
+ [monobuild](https://github.com/charypar/monobuild)
+ [transplant](https://github.com/codeactual/transplant)
+ [modmerge](https://github.com/brendanjryan/modmerge)

```shell script
git rev-parse --show-toplevel
git log -n 1 --merges --pretty=format:%p 
# -n number, --max-count=<number>
#      Limit the number of commits to output.
# --merges
#      Print only merge commits. This is exactly the same as --min-parents=2.
# --pretty[=<format>], --format=<format>
#          Pretty-print the contents of the commit logs in a given format, where <format> can be one of oneline,
#          short, medium, full, fuller, email, raw and format:<string>. See the "PRETTY FORMATS" section for
#          some additional details for each format. When omitted, the format defaults to medium.
# '%p' abbreviated parent hashes
```

### Why Monorepo? Why not?

+ [Dan Luu - Advantages of monorepos](https://danluu.com/monorepo/)
+ [Matt Klein - Monorepos: Please don’t!](https://medium.com/@mattklein123/monorepos-please-dont-e9a279be011b)
+ [Adam Jacob - Monorepos: Please do!](https://medium.com/@adamhjk/monorepo-please-do-3657e08a4b70)
+ [Getting to Know Monorepo](https://www.strv.com/blog/getting-to-know-monorepo-engineering)

> The default behavior of a polyrepo is isolation — that’s the whole point. The default behavior of a monorepo is shared responsibility and visibility — that’s the whole point. 

As an engineer, just trying to get things done in the short term, a monorepo sucks, and polyrepos rule.
As a manager, does causing friction force conversations and promote a better long term result?

In a Multirepo layout, each project has its own Version Control System (e.g. git), deployment, configuration etc. Some benefits of this approach are:

+ Teams or individuals can work independently. Each repository is autonomous and can grow separately.
+ Onboarding new developers is easier. The smaller the codebase, the easier it is to get familiar with it.
+ Simpler tooling. Projects typically do one thing and use a single language, so tooling - CI/CD and configuration - is as simple as it gets.

Monorepos benefits:

+ Easier to update shared code. Updating a common piece of code and all the places where it is used can be done in a single step - a single commit. This operation is known as an atomic commit.
+ Better discoverability. Monorepos gives us a 360° view of the codebase. This makes actions like estimating the impact of a change much easier to do. You can also leverage tools such as a global Find & Replace for quick refactors.
+ Development culture. Sharing knowledge becomes easier when everyone is working on the same codebase and can talk about the code at the same level. Growing pains are shared, helping promote a sense of empathy among the team.
+ Flexible architecture. The process of isolating (or merging) code in a Monorepo is less costly than in a Multirepo. There's no need to set up a whole new repository, configuration and deployment. This makes it easier to revert bad abstractions or over-engineered parts of a system.

### Big tools for big companies based on Google's Blaze
+ [Google - Bazel](https://github.com/bazelbuild/bazel)
+ [Facebook - Buck](https://github.com/facebook/buck)
+ [Twitter - Pants](https://github.com/pantsbuild/pants) - in python
+ [Thoughtworks - Please](https://github.com/thought-machine/please)
+ [Buildifier for BUILD language](https://github.com/bazelbuild/buildtools)