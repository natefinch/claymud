actor.WriteString("You say \"shazam!\"\nA tornado rises up around you, obscuring your vision.\nAs the tornado clears, you find yourself somewhere else.\n\n")
around(actor, actor.Name() + " says \"shazam!\"\nA tornado rises up around " + actor.Gender().Xim + ".\nWhen the tornado clears, " + actor.Gender().Xe + " is gone.\n")
actor.Relocate(location.Exits[0].Destination)
