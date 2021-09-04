# willbeason/wikipedia

This repository is a collection of code I use for analyzing wikipedia. The main
idea here is to document methods for various analysis I've done.

Generally, code starts in commands under `cmd/` and gets migrated to `pkg/` when
I figure out how to make it more general.

I make no guarantees on the interoperability of the code here, or that output
will remain consistent with time. For now these are fun, non-professional
projects.

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

A brief description of the commands so far, roughly in the order you'd use them.

`extract-wikipedia` operates on the `pages-articles-multistream` dump of
Wikipedia, extracting each element into its own file. As of this writing,
extraction results in about 213,000 files.

`clean-wikipedia` operates on the output of `extract-wikipedia`, removing noisy
elements present in Wikipedia such as tables or various non-shown elements. This
results in an intermediate set of data which is essentially "the text of
Wikipedia as a user would see it". I keep this as I'm still updating
`normalize-wikipedia`, and `extract-wikipedia` takes annoyingly long so it's
worth saving the output of this intermediate step.

`normalize-wikipedia` operates on the output of `clean-wikipedia`, performing
operations like making everything lowercase, identifying numbers and dates,
removing tables, and the like. This is the data set I directly operate on.

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
