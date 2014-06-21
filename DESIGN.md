NateMUD Design Notes
====================

## Locks

Many entities in the MUD have locks, anything with a lock also has an Id.  We
always lock the entities in Id order, to prevent deadlocks.

## DB 

NateMUD uses BoltDB to store data.  This removes any dependency on an outside
application to store data.

Entities from the game are stored in the db using encoding/gob.

### Players

Most player data aside from inventory is loaded into memory on login.  Inventory
is always accessed through the DB to ensure synchronicity.  XP is saved to the
DB at regular intervals (TBD.... possibly 1 minute).

### Locations

Permanent Location data such as room name, description, and exits is always
loaded directly from the DB (since many rooms will never be seen by players).

Ephemeral Location data such as what players, mobs, and items are in a Location
is stored in memory.

