# Package Feeds

This repo contains a few subprojects to aid in the analysis of open source packages, in particular to look for malicious software.

These are:

[Feeds](./feeds/) to watch package registries (PyPI, NPM, etc.) for changes to packages
and to make that data available via a single standard interface.

[Analysis](./analysis/) to collect package behavior data and make it available publicly
for researchers.

[Scheduler](./scheduler/) to create jobs for Analysis based on the data from Feeds.

The goal is for all of these components to work together and provide extensible, community-run
infrastructure to study behavior of open source packages and to look for malicious software.
We also hope that the components can be used independently, to provide package feeds or runtime
behavior data for anyone interested.

# Contributing

If you want to get involved or have ideas you'd like to chat about, we discuss this project in the [OSSF Securing Critical Projects Working Group](https://github.com/ossf/wg-securing-critical-projects) meetings.

See the [Community Calendar](https://calendar.google.com/calendar?cid=czYzdm9lZmhwNWk5cGZsdGI1cTY3bmdwZXNAZ3JvdXAuY2FsZW5kYXIuZ29vZ2xlLmNvbQ) for the schedule and meeting invitations.
