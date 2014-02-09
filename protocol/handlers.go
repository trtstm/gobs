package protocol

import (
	"log"
)

func (c *Connection) handlePlogin(msg Plogin) {
	if msg.Flag {
		if c.Biller.UserExists(msg.Name) {
			answer := Pbad{msg.Pid, false, "user already exists"}
			c.Send(answer)
			return
		}

		err := c.Biller.CreateUser(msg.Name, msg.Pw)
		if err != nil {
			log.Printf("(Connection::handlePlogin) could not create user: %s\n", err)
			answer := Pbad{msg.Pid, false, "could not create user"}
			c.Send(answer)
			return
		}

		log.Printf("(Connection::handlePlogin) created user: %s\n", msg.Name)
		err, _ = c.Biller.Login(msg.Name, msg.Pw, c.Zone, msg.Pid)
		if err != nil {
			answer := Pbad{msg.Pid, false, ""}
			c.Send(answer)
			return
		}		

		answer := Pok{msg.Pid, "", msg.Name, "some squad", 123, 60, "1-2-1999 6:13:35"}
		c.Send(answer)
		return
	}

	err, register := c.Biller.Login(msg.Name, msg.Pw, c.Zone, msg.Pid)
	if err != nil {
		answer := Pbad{msg.Pid, register, ""}
		c.Send(answer)
		return
	}

	answer := Pok{msg.Pid, "", msg.Name, "some squad", 123, 60, "1-2-1999 6:13:35"}
	c.Send(answer)
}

func (c *Connection) handlePenterArena(msg PenterArena) {
	billerId, ok := c.Biller.PidToBillerId(c.Zone, msg.Pid)
	if !ok {
		log.Printf("handlePenterArena: Could not lookup billerId: %d\n", msg.Pid)
		return
	}

	c.Biller.EnterArena(billerId)
}

func (c *Connection) handlePleave(msg Pleave) {
	billerId, ok := c.Biller.PidToBillerId(c.Zone, msg.Pid)
	if !ok {
		log.Printf("handlePleave: Could not lookup name: %d\n", msg.Pid)
		return
	}

	c.Biller.LeaveArena(billerId)
}
