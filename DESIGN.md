ClayMUD Design Notes
====================

ClayMUD is intended to be extremely configurable with almost zero programming
knowledge.  Configuration files use the user-friendly
[toml](https://github.com/toml-lang/toml) configuration language.

## Goroutines

### Players

Connections have their own goroutine, which ends up being one per in-game
player.  These do the command parsing.

### Workers

Each zone has its own goroutine "worker", plus there's a global worker.  Each
worker contains an event channel protected by a "gate" that is a readwrite lock.
Players send "local" events (events that only affect things in the same zone as
them) to the worker for the zone they're in.  They readlock and then read-unlock
the gate mutex before sending on the channel, ensuring that if they gate is
closed they will block before the send.  Every "tick" (100ms), the worker wakes
up, write-locks the gate (shutting out any new events) and then drains the
channel of events.  Once finished, it unlocks the gate and sleeps until the next
tick.  This gate ensures that a fast player can't get a second event on the
worker queue while the worker is consuming from the queue (otherwise the queue
might never empty).

The workers ensure that all writes to global state are synchronized without race
conditions or too much lock contention.

## DB 

ClayMUD uses BoltDB to store data.  This removes any dependency on an outside
application to store data.

Entities from the game are stored in the db using json encoded bytes.
### Users

User passwords are stored as bcrypt hashes in the db, keyed by username.  The
bcrypt cost is configurable by the administrator.

### Characters

Characters (called Players in the code) are stored in the DB with their lowercased name as
the key, ensuring we don't have duplicate names that look similar.

## Locations, Mobs, Items

Permanent Location, mob, and item data  such as name, description, etc are
loaded as json from files on disk at startup.  Even with thousands of rooms and
mobs, the memory overhead is trivial.

Ephemeral Location data such as what players, mobs, and items that are in a
Location is stored in memory.

## Scripting

ClayMUD supports extensive, dynamic scripting via an embedded Python dialect
known as [starlark](https://github.com/google/starlark-go).

