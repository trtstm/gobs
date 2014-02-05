package protocol

import (
	"log"
)

func (c *Client) handlePlogin(msg Plogin) {
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

	// Check credentials of user...
	answer := Pok{msg.Pid, "", msg.Name, "some squad", 123, 60, "1-2-1999 6:13:35"}
	c.Send(answer)
}
