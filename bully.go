package bully

import "fmt"

// msgType represents the type of messages that can be passed between nodes
type msgType int

// customized messages to be used for communication between nodes
const (
	ELECTION msgType = iota
	OK
	LEADER
	CHECKLEADER
	// TODO: add / change message types as needed
)

// Message is used to communicate with the other nodes
// DO NOT MODIFY THIS STRUCT
type Message struct {
	Pid   int
	Round int
	Type  msgType
}

// Bully runs the bully algorithm for leader election on a node
func Bully(pid int, leader int, checkLeader chan bool, comm map[int]chan Message, startRound <-chan int, quit <-chan bool, electionResult chan<- int) {

	// TODO: initialization code
	leaderCheck := false
	leaderCheckCount := 0
	iAmLeaderCheck := false
	iAmLeaderCount := 0
	leaderIsAlive := true
	for {
		// quit / crash the program
		if <-quit {
			fmt.Println(pid, "stops working..")
			leaderCheck = false
			leaderCheckCount = 0
			iAmLeaderCheck = false
			iAmLeaderCount = 0
			leaderIsAlive = false
			return
		}

		roundNum := <-startRound
		fmt.Println("MyID: ", pid, roundNum, leader)

		recievingMessages := getMessages(comm[pid], roundNum-1)
		// node coming back alive
		if leader == pid && leaderIsAlive {
			leaderIsAlive = false
			fmt.Println("node is back alive... ", pid)
			for key := range comm {
				msg := &Message{
					Pid:   pid,
					Round: roundNum,
					Type:  LEADER}
				comm[key] <- *msg
			}
			//electionResult <- leader
		}

		for _, msg := range recievingMessages {
			if msg.Type == CHECKLEADER {
				okMsg := &Message{
					Pid:   pid,
					Round: roundNum,
					Type:  OK}
				comm[msg.Pid] <- *okMsg
			} else if msg.Type == OK {
				fmt.Println("Recieved OK Message by: ", pid, " from: ", msg.Pid)
				iAmLeaderCheck = false
				iAmLeaderCount = 0
			} else if msg.Type == ELECTION {
				fmt.Println("Recieved Election Message by: ", pid, " from: ", msg.Pid)
				okMsg := &Message{
					Pid:   pid,
					Round: roundNum,
					Type:  OK}
				comm[msg.Pid] <- *okMsg
				// ask nodes who has higher id than me

				iAmLeaderCheck = true
				for key := range comm {
					if key > pid {
						msg := &Message{
							Pid:   pid,
							Round: roundNum,
							Type:  ELECTION}
						comm[key] <- *msg
					}
				}

			} else if msg.Type == LEADER {
				fmt.Println("New Leader", msg.Pid)
				iAmLeaderCheck = false
				electionResult <- msg.Pid
			}
		}

		// THE NODE WHO INITATED THE ALGO THE NODE ITSLEF IS THE NEW LEADER
		if iAmLeaderCheck && iAmLeaderCount >= 2 {
			fmt.Println("Am the new leader", pid)
			leaderCheck = false
			leaderCheckCount = 0

			for key := range comm {
				msg := &Message{
					Pid:   pid,
					Round: roundNum,
					Type:  LEADER}
				comm[key] <- *msg
			}
		}

		if iAmLeaderCheck {
			iAmLeaderCount++
		}

		// CHECKING LEADER IS ALIVE OR NOT
		if len(checkLeader) > 0 {
			fmt.Println("checkLeader", pid)
			<-checkLeader
			leaderCheck = true
			// checkLeader by sending leader a message
			msg := &Message{
				Pid:   pid,
				Round: roundNum,
				Type:  CHECKLEADER}
			comm[leader] <- *msg
		}

		if leaderCheckCount == 2 && leaderCheck {

			leaderCheckCount = 0
			leaderCheck = false
			fmt.Println("Leader Dead initiaing algo", pid)

			iAmLeaderCheck = true
			// leader is ded send message to all above
			for key := range comm {
				if key > pid {
					msg := &Message{
						Pid:   pid,
						Round: roundNum,
						Type:  ELECTION}
					comm[key] <- *msg
				}
			}
			iAmLeaderCount++
		}
		// increase leaderCheckCount for 2 round check
		if leaderCheck {
			leaderCheckCount++
		}
		// TODO: bully algorithm code
	}
}

// assumes messages from previous rounds have been read already
func getMessages(msgs chan Message, round int) []Message {
	var result []Message
	// check if there are messages to be read
	if len(msgs) > 0 {
		var m Message
		// read all messages belonging to the corresponding round
		for m = <-msgs; m.Round == round; m = <-msgs {
			result = append(result, m)
			if len(msgs) == 0 {
				break
			}
		}

		// insert back message from any other round
		if m.Round != round {
			msgs <- m
		}
	}
	return result
}

// TODO: helper functions
