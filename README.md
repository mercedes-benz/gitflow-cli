# Gitflow-CLI

[![build](https://github.com/mercedes-benz/gitflow-cli/actions/workflows/build.yml/badge.svg)](https://github.com/mercedes-benz/gitflow-cli/actions/workflows/build.yml)
[![blackduck](https://github.com/mercedes-benz/gitflow-cli/actions/workflows/blackduck.yml/badge.svg)](https://github.com/mercedes-benz/gitflow-cli/actions/workflows/blackduck.yml)

Gitflow is a [branching model](https://nvie.com/posts/a-successful-git-branching-model/) that organizes feature development, 
releases, and hotfixes into dedicated branches, providing a structured approach to managing complex software projects with
reliable releases based on [semantic versioning](https://semver.org/).

The **gitflow-cli** automates this release workflow process, saving time and reducing the risk of errors. 
It maintains a clean and consistent Git graph, contributing to overall project stability.

<img src=".github/assets/gitflow-cli-demo.png" alt="gitflow-cli-demo" width="600" />

## Installation

From within the project directory the **gitflow-cli** can be built, run and installed.

1. **Clone the repository:**

    ```bash
    git clone https://github.com/mercedes-benz/gitflow-cli.git
    cd gitflow-cli
    ```

2. **Install the application:**

   To install and run the application, use the following commands:

   ```bash
   go install
   gitflow-cli --help
   ```

   **Note:** Make sure you have [Go](https://go.dev/doc/install) installed and that the `go/bin` directory is part of your PATH.

## Usage

Before starting to use **gitflow-cli**, navigate to the Git repository you want to operate on.
Make sure the repository meets all [preconditions](#preconditions).

### Release:

To initiate a new `release/x.y.z` branch from `develop`, use the following command:

   ```bash
   gitflow-cli release start
   ```

You can now use the `release/x.y.z` branch for bug fixing, creating the release changelog, 
or even deploying your product on a staging environment. Once the release is ready, finish it with:

   ```bash
   gitflow-cli release finish
   ```

### Hotfix:

To initiate a new `hotfix/x.y.z` branch from `main`, use the following command:

   ```bash
   gitflow-cli hotfix start
   ```

Check out the `hotfix/x.y.z` branch, create a quick patch, and push your changes. Then, finish the hotfix with:

   ```bash
   gitflow-cli hotfix finish
   ```

## Preconditions

To use **gitflow-cli**, ensure your project meets the basic structural requirements, particularly around Git branches and version management.

### Git Branches

Your repository must define a dedicated **production** and **development** branches (e.g. main and develop).
These can be [customized](#configuration) as needed.

### Version File

Each project type may store version information in a different location.
The **gitflow-cli** detects your project's context and automatically delegates tasks to the appropriate plugin based on the presence of specific files.

#### Available Plugins

| Plugin       | Description                                                | Required File  | Status                                                             |
|--------------|------------------------------------------------------------|----------------|--------------------------------------------------------------------|
| **standard** | Plugin for projects without a predefined technology stack. | `version.txt`  | ![implemented](https://img.shields.io/badge/implemented-darkgreen) |
| **maven**    | Plugin for [maven](https://maven.apache.org) projects.     | `pom.xml`      | ![implemented](https://img.shields.io/badge/implemented-darkgreen) |
| **npm**      | Plugin for [npm](https://www.npmjs.com/) projects.         | `package.json` | ![planned](https://img.shields.io/badge/planned-yellow)            |

**Note:** If no technology-specific plugin can be applied, **gitflow-cli** will create a `version.txt` file in your project's root directory and apply the **standard** plugin.

## Configuration

   You have the option to provide a configuration file to **gitflow-cli**.
   This configuration file will be automatically located at `HOME/.gitflow-cli.yaml` and has the following structure:

   ```yaml
   core:
     production: main | custom-name                        # production branch name
     development: develop | custom-name                    # development branch name
     release: release | custom-name                        # release branch prefix
     hotfix: hotfix | custom-name                          # hotfix branch prefix
     undo: true | false                                    # rollback local changes in case of an error, default = false
     logging: stderr | stdout | cmdline | output | off     # diagnostic logging for the Gitflow workflow, default = stdout | cmdline | output
   ```

   You can also specify a custom configuration file using the top-level flag `--config file-path`.

## Contributing

We welcome any contributions.
If you want to contribute to this project, please read the [contributing guide](CONTRIBUTING.md).

### Git Hook

To contribute to **gitflow-cli**, we suggest setting up the Git hook below to comply with our contribution guidelines.

   ```bash
   cp .githooks/prepare-commit-msg .git/hooks/
   chmod +x .git/hooks/prepare-commit-msg
   ```

## Code of Conduct

Please read our [Code of Conduct](CODE_OF_CONDUCT.md) as it is our base for interaction.

## License

This project is licensed under the [MIT LICENSE](LICENSE).

## Provider Information

Please visit <https://www.mercedes-benz-techinnovation.com/en/imprint/> for information on the provider.

Notice: Before you use the program in productive use, please take all necessary precautions,
e.g. testing and verifying the program with regard to your specific use.
The source code has been tested solely for our own use cases, which might differ from yours.
