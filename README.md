ClayMUD
===============

ClayMUD is a highly configurable, highly performant MUD implemented in the Go programming language.

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

The game is functioning to the point where you can connect and walk around,
talk, and social.  You can configure the socials and the pronouns used for gender
(e.g. he his her).


To build and run
-----------------------

```shell
go get github.com/natefinch/claymud
claymud 
```

This will run the mud on port 8888 of your current machine. To change the port,
use -port <port>


License
-------------

MIT License
