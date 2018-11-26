``` 
 .d8888b.  888                   888b     d888 888     888 8888888b.
d88P  Y88b 888                   8888b   d8888 888     888 888  "Y88b
888    888 888                   88888b.d88888 888     888 888    888
888        888  8888b.  888  888 888Y88888P888 888     888 888    888
888        888     "88b 888  888 888 Y888P 888 888     888 888    888
888    888 888 .d888888 888  888 888  Y8P  888 888     888 888    888
Y88b  d88P 888 888  888 Y88b 888 888   "   888 Y88b. .d88P 888  .d88P
 "Y8888P"  888 "Y888888  "Y88888 888       888  "Y88888P"  8888888P"
                             888
                        Y8b d88P
                         "Y88P"
```

ClayMUD is a highly configurable, highly performant
[MUD](https://en.wikipedia.org/wiki/MUD) implemented in the Go programming
language.

Currently it is in active development, but the overarching premise is that a MUD
should be configurable and runnable without any programming knowledge. Too many
MUD systems require you to write code to change how they work. Not only is that
very likely to introduce bugs, it also restricts MUDs to be run by people who
know how to code.

ClayMUD is intended to be fully configurable through text files - everything
from the name of the MUD, to what socials are available, to how ability scores
and powers work, so that anyone can create their own unique game.


Status
-----------

The game is functioning to the point where you can connect, create an account,
create one or more players, and walk around, talk, and social.  You can
configure the socials and the pronouns used for gender (e.g. he his her).  There
is a "chatmode" toggle that will switch the interface from commands-by-default
(typical MUD), and talk-by-default (like slack or discord, where typing produces
textual output).  While in chatmode, directional commands will work normally,
but any other command must be prefixed with a configurable prefix (like /).


To build and run
-----------------------

```shell
go get -d github.com/natefinch/claymud
claymud 
```

This will run the mud on port 8888 of your current machine. To change the port,
use -port <port>

To run with version info embedded in the binary (recommended), you'll need the
[mage](magfile.org) build tool.  From the root directory of this repo:

```shell
go get github.com/magefile/mage
mage run
```


Configuration
-----------
Claymud makes extensive use of configuration files.  Working examples live in
the `data` directory of this repo, with comments explaining how they work and
what the values mean.  The main configuration file is mud.toml.

Storage
-------------

ClayMUD stores data as json serialized documents in
[boltdb](https://github.com/boltdb/bolt), an embedded key/value store that saves
to a file on disk.


License
-------------

MIT License
