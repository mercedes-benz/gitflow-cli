# Contributing

This document explains how to contribute to this project.
By contributing you will agree that your contribution will be put under the same license as this repository.

## Table of Contents
- [Communication](#communication)
- [Contributions](#contributions)
- [Developer Certificate of Origin](#developer-certificate-of-origin)
- [Quality](#quality)

## Communication

For communication please respect our [FOSS Code of Conduct](CODE_OF_CONDUCT.md).

The following communication channels exist for this project:
- Github for reporting and claiming issues: https://github.com/mercedes-benz/gitflow-cli/issues

Transparent and open communication is important to us. 
Thus, all project-related communication should happen only through these channels and in English. 
Issue-related communication should happen within the concerned issue.

## Contributions
If you would like to contribute code you can do so through GitHub by forking the repository and sending a pull request.

When submitting code, please make every effort to follow existing conventions and style in order to keep the code as readable as possible.

If you are new to contributing in Github, [First Contributions](https://github.com/firstcontributions/first-contributions) might be a good starting point.

## Developer Certificate of Origin

Contributors must agree to the "Developer Certificate of Origin Version 1.1" (DCO) and comply with the sign-off procedure for contributing to `gitflow-cli`.
All contributions submitted to the `gitflow-cli` repository have to carry a signed-off-by statement in the commit message using the following format:

    Signed-off-by: Firstname Lastname <email.address>

where the specified email address must be associated with your [github account](https://docs.github.com/en/account-and-profile/concepts/email-addresses). If a commit does not carry such a signed-off-by statement, the pull request cannot be merged into the official `gitflow-cli` repository because by adding signed-off-by statements to the commit messages, you convey that you have the necessary rights to contribute the respective changes.

Git adds the sign-off line automatically when you commit with the `-s` (`--signoff`) flag:

```bash
git commit -s -m "fix: correct version parsing"
```

If you forgot to sign off your last commit, you can amend it:

```bash
git commit --amend -s --no-edit
```

To sign off multiple existing commits on your branch, rebase them (replace `main` with your target branch):

```bash
git rebase --signoff main
```

Pull requests are checked automatically by the DCO app; all commits must be signed off before the pull request can be merged.

### Developer Certificate of Origin (Text)
```text
Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I have the right to submit it under the open source license indicated in the file; or

(b) The contribution is based upon previous work that, to the best of my knowledge, is covered under an appropriate open source license and I have the right under that license to submit that work with modifications, whether created in whole or in part by me, under the same open source license (unless I am permitted to submit under a different license), as indicated in the file; or

(c) The contribution was provided directly to me by some other person who certified (a), (b) or (c) and I have not modified it.

(d) I understand and agree that this project and the contribution are public and that a record of the contribution (including all personal information I submit with it, including my sign-off) is maintained indefinitely and may be redistributed consistent with this project or the open source license(s) involved.
```

The text of the DCO is also available at <https://developercertificate.org>.

## Quality
Please ensure that for all contributions, the corresponding documentation is in-sync and up-to-date. All documentation should be in English language. 

We assume that for every non-trivial contribution, the project has been built and tested prior to the contribution. 
