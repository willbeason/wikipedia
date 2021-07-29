# willbeason/wikipedia

This repository is a collection of code I use for analyzing wikipedia. The main
idea here is to document methods for various analysis I've done.

Generally, code starts in commands under `cmd/` and gets migrated to `pkg/` when
I figure out how to make it more general.

I make no guarantees on the inter-operability of the code here, or that output
will remain consistent with time. For now these are fun, non-professional
projects. However, do feel free to ask if you're interested in using something
here!

Many of these commands will change with time as I learn more about the Wikipedia
corpus. This is a large, complex set of data and there aren't (yet) good
resources on the types of things these libraries need to account for.

## The Commands

A brief description of the commands so far, roughly in the order you'd use them.

`extract-wikipedia` operates on the `pages-articles-multistream` dump of
Wikipedia, extracting each element into its own file. As of this writing,
extraction results in about 213,000 files.

`clean-wikipedia` operates on the output of `extract-wikipedia`, removing noisy
elements present in Wikipedia such as tables or various non-shown elements.

`title-words` operates on the output of `clean-wikipedia`, creating a frequency
table of words present in the titles of Wikipedia articles. Requires an input
dictionary of already-known words.

`wordcount` operates on the output of `clean-wikipedia`, creating a frequency
table of words present in Wikipedia articles. Requires input dictionary of
known words, as the full set of unique words on Wikipedia is too large to hold
in memory.

`wordpresence` operates on the output of `clean-wikipedia`, checking for the
presence of words from a frequency table in each article. Writes to disk
a sorted set of the ids of the words present in each.

`colocation` operates on the output of `wordpresence`. It is configured to look
for patterns of words appearing together in the same article (or not).
