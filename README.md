# willbeason/wikipedia

This repository is a collection of code I use for analyzing wikipedia. The main
idea here is to document methods for various analysis I've done.

Generally, code starts in commands under `cmd/` and gets migrated to `pkg/` when
I figure out how to make it more general and deduplicate code across commands.

I make no guarantees on the interoperability of the code here, or that output
will remain consistent with time. For now these are fun, non-professional
projects. Generally, I will tag commits after making significant changes.
As yet there is no stable release of this software, but I am working towards
this as I make progress towards using this code for research papers.

Many of these commands will change with time as I learn more about the Wikipedia
corpus. This is a large, complex set of data and there aren't (yet) good
resources on the types of things these libraries need to account for.

## The Data

Commands in this repository mainly operate on either the regular
pages-articles-multistream Wikipedia dumps, or Badger databases constructed from
the commands in this repository. So (for now) you can't really look at
individual articles unless you know their Wikipedia Page ID (You can get this by
clicking "Page information" on any page on Wikipedia.org).

The data format is optimized for streaming the entirety of Wikipedia. I've
offloaded concerns about indexing articles or how large to make files to Badger
as this quickly became a headache I didn't want to deal with.

## The Commands

I am reworking this functionality heavily to make using individual commands
easier, and allow for workflows to be repeated.

The main command is `wikopticon`, found in `cmd/wikopticon`. You can use it
in one of two ways.

1. (Recommended) Use a configuration file that defines individual jobs which
    rely on subcommands.
2. (Not Recommended) Manually call subcommands with `wikopticon subcommand`.

## Warnings

Some commands just fail - they'll panic in bader's DB.Orchestrate. I don't have
a good explanation for this. Either there's a bug in Badger when streaming data
in 32 threads, or I'm using something wrong. In any case, rerunning the command
several times seems to work fine. Each time the process will make more progress,
and eventually it succeeds. For now commands take on the order of a couple of
minutes for  me to rerun, so it isn't worth my time to debug.

Memory consumption is essentially unbounded since I haven't messed with any of
the Badger memory settings. This is fine for my machine since I have 64 GB of
memory and can hold most (or in some cases all) of Wikipedia in memory at once,
but you may experience slowdowns if you don't modify the options set in
`pkg/db/process.go`.
