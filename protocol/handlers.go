package protocol

import (
	"log"
)

func (c *Connection) handlePlogin(msg Plogin) {
	if msg.Flag {
		if c.Biller.UserExists(msg.Name) {
			answer := Pbad{msg.Pid, false, "Can't create user. User already exists"}
			c.Send(answer)
			return
		}

		err := c.Biller.CreateUser(msg.Name, msg.Pw)
		if err != nil {
			log.Printf("handlePlogin: Could not create user: %s\n", err)
			answer := Pbad{msg.Pid, false, "Could not create user"}
			c.Send(answer)
			return
		}

		log.Printf("handlePlogin: Created user: %s\n", msg.Name)
		answer := Pok{msg.Pid, "", msg.Name, "some squad", 123, 60, "1-2-1999 6:13:35"}
		c.Send(answer)
		return
	}

	if !c.Biller.UserExists(msg.Name) {
		answer := Pbad{msg.Pid, true, "User does not exist"}
		c.Send(answer)
		return
	}

	if c.Biller.LoggedIn(msg.Name) {
		answer := Pbad{msg.Pid, false, "Already logged in"}
		c.Send(answer)
		return
	}

	err, register := c.Biller.Login(msg.Name, msg.Pw)
	if err != nil {
		answer := Pbad{msg.Pid, register, ""}
		c.Send(answer)
		return
	}

	answer := Pok{msg.Pid, "", msg.Name, "some squad", 123, 60, "1-2-1999 6:13:35"}
	c.Send(answer)
}

func (c *Connection) handlePenterArena(msg PenterArena) {
	name, ok := c.Biller.PidToName(c.Zone, msg.Pid)
	if !ok {
		log.Printf("handlePenterArena: Could not lookup name: %d\n", msg.Pid)
		return
	}

	c.Biller.EnterArena(name, c.Zone)
}

func (c *Connection) handlePleave(msg Pleave) {
	name, ok := c.Biller.PidToName(c.Zone, msg.Pid)
	if !ok {
		log.Printf("handlePleave: Could not lookup name: %d\n", msg.Pid)
		return
	}

	c.Biller.LeaveArena(name, c.Zone)
}
