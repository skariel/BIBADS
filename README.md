# BIBADS
This program generates a bib file to be used with bibtex. The name of the generated bib file is as needed by the tex file, for e.g. if the tex file contains `\bibliography{KN16.bib}` the
generated file name will be `KN16.bib`. Beware this will overwrite any existing file with the same name and path.

The generated bib file will contain all the citations needed by the tex file that can be found in the Nasa Astrophysics Data System (Nasa ADS)

All citations in the tex file need to reference bibcodes, e.g. `\cite{2009MNRAS.399..683J}`

Aliases are optional, and are defined inside a comment in the tex file, for e.g.:

```
% bibalias colles2DF 2001MNRAS.328.1039C
% bibalias peeb80    1980lssu.book.....P
```

Then the citation can look like `\cite{peeb80, colles2DF}`

# RUN THE PROGRAM

One option is to run it as a script, this needs a go installation (google golang):

```
> go run bibads.go [your tex file here]
```

A second option is compiling, this need a go installation, it generates a self-contained executable:

```
> go build bibads.go
```

You can also use a precompiled binary for your OS, there currently is one for OSX, Linux and Windows (all for amd64 arch). When you have a working binary just run:

```
./bibads [your tex file here]
```

Note that the program will not fetch from the ADS any entries already present in the bib file. This is known as caching. If you want to update (and overwrite) all entries, a `-nocache` flag should be used as a first parameter:

```
> ./bibads -nocache mypaper.tex
```

# TODO
better command line interface (CLI) i.e. better help, handle different parameter ordering, show version, etc.

# MOTIVATION
1. When using this program there is not need to maintain a separate bib file. The bib file is very personal, sometimes it is managed by "knowledge" systems, it is hard to share, merge, etc.
2. Updating all bib entries is as easy as running a single command

# WHY GO?
1. It is statically typed, which I like
2. It compiles fast, I like this too
3. Cross compile to Linux, OSX and Windows... I like!
4. Compiles to self contained binaries, no need to install a thing in this mode
5. Can run program as a script
6. Efficient (fast, etc.)
7. Easy concurrency/parallelism/asynchronicity

# LICENSE
do whatever you want with this, use on your own risk

# AUTHOR
Ariel Keselman

# VERSION
1.2.0 2016

