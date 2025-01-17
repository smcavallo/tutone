<a name="v0.10.1"></a>
## [v0.10.1] - 2021-09-27
### Bug Fixes
- Add release info to README

<a name="v0.10.0"></a>
## [v0.10.0] - 2021-09-27
### Features
- enable auto-releases

<a name="v0.9.0"></a>
## [v0.9.0] - 2021-09-15
### Features
- add skip_fields for skipping fields within a type
- add custom template funcs
- **generator:** add ability to override struct tags

<a name="v0.8.1"></a>
## [v0.8.1] - 2021-06-15
### Bug Fixes
- **schema:** Explicitly anchor the regexp for mutation name (MatchString does not)

<a name="v0.8.0"></a>
## [v0.8.0] - 2021-06-15
### Features
- **schema:** Allow for mutations to be matched by regexp instead of statically declared in config

<a name="v0.7.0"></a>
## [v0.7.0] - 2021-06-14
### Bug Fixes
- **fetch:** allow plain graphql endpoints for local development

### Features
- **schema:** Add ability to filter out specific fields in queries/mutations

<a name="v0.6.1"></a>
## [v0.6.1] - 2021-02-11
### Bug Fixes
- **golang:** Use Golang field names, not title case of path
- **nerdgraphclient:** Return value types need name overrides

### Documentation Updates
- Fix repository URL in changelog

<a name="v0.6.0"></a>
## [v0.6.0] - 2021-01-27
### Bug Fixes
- **schema:** Queries withouth args do not get ()

<a name="v0.5.0"></a>
## [v0.5.0] - 2021-01-04
### Features
- **golang:** Add ability to generate Get funcs for structs

<a name="v0.4.3"></a>
## [v0.4.3] - 2020-12-21
### Bug Fixes
- **schema:** Do not expand fields of types we will not create
- **schema:** The endpoint for a query might not take args

<a name="v0.4.2"></a>
## [v0.4.2] - 2020-12-16
<a name="v0.4.1"></a>
## [v0.4.1] - 2020-12-15
<a name="v0.4.0"></a>
## [v0.4.0] - 2020-12-15
### Bug Fixes
- **config:** Mutation and Query field depth params should match
- **schema:** improve Mutation query generation

### Documentation Updates
- Add some info on templates

### Features
- **schema:** Enable mutations to have fields required in GraphQL (work-around schema issues)

<a name="v0.3.0"></a>
## [v0.3.0] - 2020-12-04
### Bug Fixes
- **golang:** Missing 0 == more allocations...
- **golang:** interface method memory usage can be right-sized
- **golang:** pass input prereqs for method signature
- **golang:** skip_type_create should skip for all types, not just scalars
- **golang:** ensure prefix for method arguments
- **schema:** ensure proper casing of mutation names
- **schema:** reduce handling of query args to only non-null
- **schema:** ensure proper handling of query args
- **typegen:** Avoid panic on nil pointer unmarshal

### Features
- **command:** use schema types to build CLI command
- **command:** add ability to generate READ commands (amend this commit with cleanup)
- **format:** programmatically run goimports on generated code, template updates
- **query:** Enable nullable fields in query (this enables all of them...)
- **schema:** implement helper to gatehr input variables from path
- **typegen:** Add package import path to config
- **typegen:** Allow custom methods to be added to an interface definition

<a name="v0.2.5"></a>
## [v0.2.5] - 2020-10-08
### Bug Fixes
- **build:** update changelog action for improved standards
- **deps:** use v3 package for sprig
- **schema:** use better comparison when overriding field names

<a name="v0.2.4"></a>
## [v0.2.4] - 2020-10-07
### Bug Fixes
- **expander:** ensure expansion of type arguments
- **golang:** ensure list kinds are represented as slices
- **golang:** ensure more sorting for deterministic output
- **golang:** sort the methods before return
- **schema:** ensure possibleTypes on interfaces are expanded

### Documentation Updates
- fix casing in README and format
- tidy up on some documentation

### Features
- add new generator for generating CLI commands
- include query string handling for golang
- begin method to build a query string from a Type

<a name="v0.2.3"></a>
## [v0.2.3] - 2020-09-04
<a name="v0.2.2"></a>
## [v0.2.2] - 2020-09-03
### Bug Fixes
- **golang:** move Interface reference to template
- **release:** update project name for goreleaser

<a name="v0.2.1"></a>
## [v0.2.1] - 2020-09-03
### Bug Fixes
- **changelog:** update changelog on release only, drop reviewer spec
- **golang:** ensure Name references use goFormatName()
- **schema:** use correct name for lookups

<a name="v0.2.0"></a>
## [v0.2.0] - 2020-09-02
### Bug Fixes
- ensure only specific package types are generated when passing --package option
- **codegen:** update package ref for go mod usage
- **generate:** ensure correct generator client
- **lang:** remove pointer reference from return type
- **nerdgraphclient:** move condition block end to exclude mutation
- **schema:** avoid recursing forever when handling interface kinds
- **schema:** ensure recursive expansion for additional Kinds
- **schema:** ensure proper handling of list interfaces

### Features
- **codegen:** implement sprig community template functions
- **lang:** begin GoMethod implementation
- **schema:** implement type expansion based on method name

<a name="v0.1.2"></a>
## [v0.1.2] - 2020-08-14
<a name="v0.1.1"></a>
## [v0.1.1] - 2020-07-23
### Bug Fixes
- **schema:** avoid expanding a type twice

<a name="v0.1.0"></a>
## v0.1.0 - 2020-07-23
### Bug Fixes
- **fetch:** exit non-zero on fatal log message
- **generate:** dont double prepend [] for list types - i.e. [][]type
- **generate:** remove generate.yml and all instances of it's reference
- **schema:** ensure proper handling of LIST types
- **util:** ensure fields of nested types are also expanded

### Documentation Updates
- **tutone:** include a couple doc strings
- **tutone:** include a what? section

### Features
- **fetch:** Fetch root mutation type
- **fetch:** Generic schema fetching and cache to file
- **generate:** format the generated source code according Go standards
- **generate:** WIP - first attempt at getting tutone to generate types
- **generate:** Generate command reading configs
- **generate:** implement --refetch flag
- **generate:** fetch if schema not present
- **generator:** introduce a generator concept
- **tutone:** default path for tutone config file

[Unreleased]: https://github.com/newrelic/tutone/compare/v0.10.1...HEAD
[v0.10.1]: https://github.com/newrelic/tutone/compare/v0.10.0...v0.10.1
[v0.10.0]: https://github.com/newrelic/tutone/compare/v0.9.0...v0.10.0
[v0.9.0]: https://github.com/newrelic/tutone/compare/v0.8.1...v0.9.0
[v0.8.1]: https://github.com/newrelic/tutone/compare/v0.8.0...v0.8.1
[v0.8.0]: https://github.com/newrelic/tutone/compare/v0.7.0...v0.8.0
[v0.7.0]: https://github.com/newrelic/tutone/compare/v0.6.1...v0.7.0
[v0.6.1]: https://github.com/newrelic/tutone/compare/v0.6.0...v0.6.1
[v0.6.0]: https://github.com/newrelic/tutone/compare/v0.5.0...v0.6.0
[v0.5.0]: https://github.com/newrelic/tutone/compare/v0.4.3...v0.5.0
[v0.4.3]: https://github.com/newrelic/tutone/compare/v0.4.2...v0.4.3
[v0.4.2]: https://github.com/newrelic/tutone/compare/v0.4.1...v0.4.2
[v0.4.1]: https://github.com/newrelic/tutone/compare/v0.4.0...v0.4.1
[v0.4.0]: https://github.com/newrelic/tutone/compare/v0.3.0...v0.4.0
[v0.3.0]: https://github.com/newrelic/tutone/compare/v0.2.5...v0.3.0
[v0.2.5]: https://github.com/newrelic/tutone/compare/v0.2.4...v0.2.5
[v0.2.4]: https://github.com/newrelic/tutone/compare/v0.2.3...v0.2.4
[v0.2.3]: https://github.com/newrelic/tutone/compare/v0.2.2...v0.2.3
[v0.2.2]: https://github.com/newrelic/tutone/compare/v0.2.1...v0.2.2
[v0.2.1]: https://github.com/newrelic/tutone/compare/v0.2.0...v0.2.1
[v0.2.0]: https://github.com/newrelic/tutone/compare/v0.1.2...v0.2.0
[v0.1.2]: https://github.com/newrelic/tutone/compare/v0.1.1...v0.1.2
[v0.1.1]: https://github.com/newrelic/tutone/compare/v0.1.0...v0.1.1
