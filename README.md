<p align="center"><h1 align="center">Sqlt</h1></p>
<p align="center">
	<em><code>❯ SQL Templating</code></em>
</p>
<p align="center">
	<img src="https://img.shields.io/github/license/hxtk/sqlt?style=default&logo=opensourceinitiative&logoColor=white&color=0080ff" alt="license">
	<img src="https://img.shields.io/github/last-commit/hxtk/sqlt?style=default&logo=git&logoColor=white&color=0080ff" alt="last-commit">
	<img src="https://img.shields.io/github/languages/top/hxtk/sqlt?style=default&color=0080ff" alt="repo-top-language">
	<img src="https://img.shields.io/github/languages/count/hxtk/sqlt?style=default&color=0080ff" alt="repo-language-count">
    <a href="https://pkg.go.dev/github.com/hxtk/sqlt"><img src="https://pkg.go.dev/badge/github.com/hxtk/sqlt.svg" alt="Go Reference"></a>
</p>
<p align="center"><!-- default option, no dependency badges. -->
</p>
<p align="center">
	<!-- default option, no dependency badges. -->
</p>
<br>

##  Table of contents

- [Overview](#overview)
  - [Example](#example)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Usage](#usage)
  - [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgments](#acknowledgments)

---

##  Overview

Sqlt provides safe template-based SQL generation for Golang.

This library primarily exists to solve the human problem of code review for
generated SQL queries. This allows users to write their SQL queries in SQL
files, where `CODEOWNERS` scanning can easily route changes affecting SQL query
generation to database SMEs who can then easily see the structure of the
resulting queries without needing to sift through branches of Go code in which
they may have less expertise reading.

It does so by using a wrapper around Golang's standard library `text/template`
templating package that escapes templates by ensuring any Action Nodes-the
parts of the template that render runtime data-end their pipeline with a safe
SQL sanitizer function.

Users may provide their own bindings of SQL sanitizers and end action node
pipelines with those custom SQL sanitizers. If a pipeline doesn't end with
a registered sanitizer, this wrapper safely escapes the value by appending
a default sanitizer that replaces the value with a named parameter binding.

### Example

In the following example, the code parses a template from a string literal for
simplicity. You may wish to use `forbidigo` or some other linter to forbid
parsing templates from string literals and require developers to use the
`ParseFS` parser, which loads templates from a filesystem-like object such as
an `embed.FS`. This ensures that every line of SQL exists in a separate file.

```go

tpl, err := New("q").Parse(`SELECT * FROM users WHERE id={{ .ID }}`)
if err != nil {
    panic(err)
}

query, args, err := tpl.Execute(map[string]any{"ID": 123})
if err != nil {
    panic(err)
}

pool, err := pgxpool.New(context.TODO(), os.Getenv("DATABASE_URL"))
if err != nil {
	panic(err)
}

rows, err := pool.Query(context.TODO(), query, args)
if err != nil {
	panic(err)
}

user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[User])
if err != nil {
	panic(err)
}

// use User in your code.
```

##  Getting started

###  Prerequisites

Before getting started with sqlt, ensure your runtime environment meets the following requirements:

- **Programming Language:** Go
- **Package Manager:** Go modules

###  Installation

Install sqlt using one of the following methods:

**Build from source:**

1. Clone the sqlt repository:
```sh
❯ git clone https://github.com/hxtk/sqlt
```

2. Navigate to the project directory:
```sh
❯ cd sqlt
```

3. Install the project dependencies:

**Using `go modules`** &nbsp; [<img align="center" src="https://img.shields.io/badge/Go-00ADD8.svg?style={badge_style}&logo=go&logoColor=white" />](https://golang.org/)

```sh
❯ go build
```

###  Testing

Run the test suite using the following command:
**Using `go modules`** &nbsp; [<img align="center" src="https://img.shields.io/badge/Go-00ADD8.svg?style={badge_style}&logo=go&logoColor=white" />](https://golang.org/)

```sh
❯ go test ./...
```

---

##  Contributing

*   **💬 [Join the Discussions](https://github.com/hxtk/sqlt/discussions)**:
    Share your insights, provide feedback, or ask questions.
*   **🐛 [Report Issues](https://github.com/hxtk/sqlt/issues)**: Submit bugs
    found or log feature requests for the `sqlt` project.
*   **💡 [Submit Pull Requests](https://github.com/hxtk/sqlt/blob/main/CONTRIBUTING.md)**: Review open PRs, and submit your own PRs.

<details closed>
<summary>Contributor Graph</summary>
<br>
<p align="left">
   <a href="https://github.com{/hxtk/sqlt/}graphs/contributors">
      <img src="https://contrib.rocks/image?repo=hxtk/sqlt">
   </a>
</p>
</details>

### Contributing code

For code contributions, this repository follows the practices described in
[Google's engineering practices handbook](https://google.github.io/eng-practices/review/developer/).
Note that in the standard of code review, the standard states that a reviewer
should approve a CL if it offers a net improvement to the codebase, even if
they wouldn't call it perfect. "A net improvement to the codebase" leaves room
for maintainer discretion when it comes to feature additions. Contributors
should open an issue first if they would like to know whether maintainers would
accept a feature before they begin working on it.

Commit messages should use
[Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

Authors can expect code review comments to use
[Conventional Comments](https://conventionalcomments.org/).

Authors shouldn't make new commits for code review fixups. In general, use the
structure of commits to tell a story about the evolution of the library. Keep
each commit reasonably small, and limit a pull request to a small number of
them. A simple pull request should only include one commit, but you might
submit a more complex change as a series of "red, green, refactor" commits
typical of test-driven development. If you update your pull request in response
to comments, do so by amending the original commit and force-pushing to your
work branch.

[Jujutsu](https://jj-vcs.github.io/jj/latest/) provides a cleaner interface for
authoring commits in the style the maintainers of this library prefer.

##  License

The authors distribute this project under the
Apache 2.0 License. For more details, refer to the [LICENSE](LICENSE) file.

##  Acknowledgments

For the developer experience and ergonomics of developing database code, this
project draws inspiration from https://github.com/sqlc-dev/sqlc. The
maintainers of this repository thank the maintainers of sqlc for inspiring the
workflow of authoring SQL code by writing SQL queries in separate files where
SQL experts have an easier time reviewing them, repository management
structures such as the CODEOWNERS file have an easier time ensuring that SQL
changes actually get reviewed by a SQL expert, and code editors have an easier
time correctly highlighting the syntax.

This project differentiates itslf from `sqlc` because it doesn't include a
build pipeline stage and it supports more dynamic queries, possibly with
user-provided SQL generators. This dynamicism comes at the expense of the
query analysis techniques that `sqlc` provides during its build phase.

For its implementation, this project draws heavy inspiration from
https://github.com/VauntDev/tqla. The maintainers thank @VauntDev for inspiring
the use of escaped `text/template`s to generate SQL.

This project differentiates itself from `tqla` by aligning more with the base
`text/template` library to support writing templates in SQL files and reduce
duplicated parsing work. It also supports "bring-your-own" SQL sanitizers
besides the default sanitizer, so that users may safely render, for example,
`WHERE` clauses generated by a safe filter string parser.

