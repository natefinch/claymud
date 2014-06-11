NateMUD
===============

NateMUD is a highly configurable, highly performant MUD implemented in the Go programming language.

Currently it is in active development, but the overarching premise is that a MUD should be configurable and runnable without any programming knowledge. Too many MUD systems require you to write code to change how they work. Not only is that very likely to introduce bugs, it also restricts MUDs to be run by people who know how to code.

NateMUD is intended to be fully configurable through text files - everything from the name of the MUD, to what emotes are available, to how ability scores and powers work, so that anyone can create their own unique game.


Status
-----------

The game is functioning to the point where you can connect and walk around, talk, and emote.  You can configure the emotes and the pronouns used for gender (e.g. he his her).


To build and run
-----------------------

```shell
go get github.com/natefinch/natemud
natemud 
```

This will run the mud on port 8888 of your current machine. To change the port, use -p <port>


License
-------------

MIT License


About Nate
----------

I have been developing software professionally since 1999.  In college in the late 90's I was extensively into MUDs, was an admin on a MUD and even ran a heavily modified Circle MUD for a time after college. I also have been playing tabletop RPGs for over 20 years. I have an extensive collection of tabletop RPGs and "German" board games that influence my thinking.  I'm approaching this project with an eye towards game design, with the hope that I can give a framework to other game developers for creating their own unique game.
