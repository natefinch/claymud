def do():
    for name, p in location.Players.items():
        if p.ID != actor.ID:
            p.WriteString(actor.Name() + " pushes a button on the wall.\n")
        p.WriteString("A cold breeze whistles through the room.\n")
        p.WriteString("A quiet voice whispers \"Hello " + actor.Name() + ".\"\n")

actor.WriteString("The button slowly depresses.\n")
do()
