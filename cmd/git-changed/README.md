### git-changed
Equivalent to `git diff --name-only`

### Notes

In addition to the [normal](https://github.com/StevenACoffman/cicd-demo) [CI/CDD concerns](https://docs.google.com/presentation/d/1h62_B_2eiuOE1iVIFpEm0C5f5MaJQ1N-zLKJM6FhxmA/edit?usp=sharing), Monorepos have additional complications.

Identifying what has changed, especially in a monorepo, is a critical first step.

If one or more source files are altered, which services must be rebuilt, retested, and redeployed? If multiple services must be deployed, what overall deployment strategy or sequence is necessary? What happens when something needs to be rolled back for only a single service?

+ Some organizations use configuration to assign specific changes to specific services/actions.
+ Some organizations introspectively identify dependency trees (this is typically language specific).
+ Some organizations use subtree splits and git submodules to provide atomic, independently deployable view of services.
+ Some organizations use monorepos for apps, but put shared code (libraries) in other repos

Monorepos pose some common problem, without a single standard solution. The different strategies that have been applied to this problem space can be seen by the variety of tools people have produced. I'm focusing on the ecosystem in Go.

|   | multirepositories | monorepository  |
|---|---|---|
| monomodules  | monomodule multi-repositories   | monomodule monorepo  |
| multimodules  | multi-module multi-repositoryies | multi-module monorepo |


##### Other Monorepo tools

+ [GTA](https://github.com/jphines/gta) - Go Test Auto - [Digital Ocean uses this](https://blog.digitalocean.com/cthulhu-organizing-go-code-in-a-scalable-repo/)
+ [detect-changed-services](https://github.com/devops-recipes/app-mono/blob/master/detect-changed-services.sh) - [Shippable uses this](http://blog.shippable.com/build-test-and-deploy-applications-independently-from-a-monorepo)
+ [golang-monorepo](https://github.com/flowerinthenight/golang-monorepo) - [Mobingi uses this](https://tech.mobingi.com/2018/09/25/ouchan-monorepo.html)
+ [Affected](https://github.com/jharlap/affected)
+ [sowhatsnew](https://github.com/manifoldco/sowhatsnew/)
+ [tainted](https://github.com/kynrai/tainted)
+ [mono-meta](https://github.com/davidae/mono-meta)
+ [knega](https://github.com/kristofferlind/knega)
+ [mona](https://github.com/uw-labs/mona)
+ [baur](https://github.com/simplesurance/baur)
+ [monorepo-operator](https://github.com/SimonBaeumer/monorepo-operator) - subtree splits
+ [mbt](https://github.com/mbtproject/mbt)
+ [drone-tree-config](https://github.com/bitsbeats/drone-tree-config) (github pr based dir diff)
+ [monobuild](https://github.com/charypar/monobuild)
+ [transplant](https://github.com/codeactual/transplant)
+ [modmerge](https://github.com/brendanjryan/modmerge)

For other monorepo things that are not Go centric, check out [Awesome Monorepo](https://github.com/korfuri/awesome-monorepo)

### Big tools for big companies based on Google's Blaze

+ [Google - Bazel](https://github.com/bazelbuild/bazel)
+ [Thoughtworks - Please](https://github.com/thought-machine/please) - in Go!
+ [Facebook - Buck](https://github.com/facebook/buck)
+ [Twitter - Pants](https://github.com/pantsbuild/pants) - in python
+ [Buildifier for BUILD language](https://github.com/bazelbuild/buildtools)

### Why Monorepo? Why not?

+ [Dan Luu - Advantages of monorepos](https://danluu.com/monorepo/)
+ [Matt Klein - Monorepos: Please don’t!](https://medium.com/@mattklein123/monorepos-please-dont-e9a279be011b)
+ [Adam Jacob - Monorepos: Please do!](https://medium.com/@adamhjk/monorepo-please-do-3657e08a4b70)
+ [Getting to Know Monorepo](https://www.strv.com/blog/getting-to-know-monorepo-engineering)
+ [Why We Should Not Return to Monolithic Repositories](https://gist.github.com/technosophos/9c706b1ef10f42014a06)

> The default behavior of a polyrepo is isolation — that’s the whole point. The default behavior of a monorepo is shared responsibility and visibility — that’s the whole point.

As an engineer, just trying to get things done in the short term, a monorepo sucks, and polyrepos rule.
As a manager, does causing friction force conversations and promote a better long term result?

In a Multirepo layout, each project has its own Version Control System (e.g. git), deployment, configuration etc. Some benefits of this approach are:

+ Teams or individuals can work independently. Each repository is autonomous and can grow separately.
+ Onboarding new developers is easier. The smaller the codebase, the easier it is to get familiar with it.
+ Simpler tooling. Projects typically do one thing and use a single language, so tooling - CI/CD and configuration - is as simple as it gets.

##### Monorepos benefits:

+ Easier to update shared code. Updating a common piece of code and all the places where it is used can be done in a single step - a single commit. This operation is known as an atomic commit.
+ Better discoverability. Monorepos gives us a 360° view of the codebase. This makes actions like estimating the impact of a change much easier to do. You can also leverage tools such as a global Find & Replace for quick refactors.
+ Development culture. Sharing knowledge becomes easier when everyone is working on the same codebase and can talk about the code at the same level. Growing pains are shared, helping promote a sense of empathy among the team.
+ Flexible architecture. The process of isolating (or merging) code in a Monorepo is less costly than in a Multirepo. There's no need to set up a whole new repository, configuration and deployment. This makes it easier to revert bad abstractions or over-engineered parts of a system.
### Managing multirepos
+ [mr - tool to manage all your version control repositories](https://myrepos.branchable.com/)
+ [github2mr - convert github repos to mr](https://github.com/skx/github2mr)
+ [ghq - Manage remote repository clones](https://github.com/x-motemen/ghq)
+ [vcs - Repository Management for Go](https://github.com/Masterminds/vcs)
+ [gitbatch](https://github.com/isacikgoz/gitbatch)
### Some more git stuff
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
### Some Build System Theory
The [Build Systems à la Carte](https://www.microsoft.com/en-us/research/publication/build-systems-la-carte/) paper proposes splitting build systems into two components:

+ Rebuilders decide when to rebuild a particular key (file).
+ Schedulers decide how to rebuild multiple keys - handling dependencies while maintaining correctness and efficiency.

Schedulers come in 3 flavors (see Section 4):

+ Topological use a simple topological sort to determine the order of building keys. They are limited to static dependencies - if new dependencies are discovered during the build process, they cannot be correctly added to the build graph.
+ Restarting schedulers start building a particular key, then, as dependencies are discovered, abort building that key. Then the discovered dependency(ies) are built, after which the original key’s task is restarted. This allows dynamic dependencies, but at the cost of repeating parts of the build process.
+ Suspending schedulers also start building a particular key. When a dependency is discovered, instead of aborting the current task, they suspend it, go and build the dependency, then resume the task. This allows dynamic dependencies without repeating work.

From [A Future is a Suspending Scheduler](https://nikhilism.com/post/2020/futures-suspending-scheduler/)
