package ws

type Hub struct {
	Register    chan *User
	Unregister  chan *User
	connections map[*User]bool
	onlineUsers map[int32][]*User
}

var DefaultHub = Hub{
	Register:    make(chan *User),
	Unregister:  make(chan *User),
	connections: make(map[*User]bool),
	onlineUsers: make(map[int32][]*User),
}

func (this *Hub) Run() {
	for {
		select {
		case user := <-this.Register:
			this.RegisterUser(user)
		case user := <-this.Unregister:
			this.UnregisterUser(user)
		}
	}
}

func (this *Hub) RegisterUser(user *User) {
	if user != nil {
		this.connections[user] = true
		if conns, ok := this.onlineUsers[user.ID]; ok {
			this.onlineUsers[user.ID] = append(conns, user)
		} else {
			this.onlineUsers[user.ID] = []*User{user}
		}
	}
}

func (this *Hub) UnregisterUser(user *User) {
	if user != nil {
		delete(this.connections, user)
		close(user.Send)
		if conns, ok := this.onlineUsers[user.ID]; ok {
			index := -1
			for i, conn := range conns {
				if conn == user {
					index = i
					break
				}
			}
			if index != -1 {
				this.onlineUsers[user.ID] = append(conns[:index], conns[index+1:]...)
			}
			if len(this.onlineUsers[user.ID]) == 0 {
				delete(this.onlineUsers, user.ID)
			}
		}
	}
}

func (this *Hub) GetUser(userID int32) []*User {
	if userID != 0 {
		if conns, ok := this.onlineUsers[userID]; ok {
			return conns
		}
	}
	return nil
}

func (this *Hub) GetUsers(usersList []int32) []*User {
	if usersList != nil {
		users := []*User{}
		for _, userID := range usersList {
			if conns, ok := this.onlineUsers[userID]; ok {
				users = append(users, conns...)
			}
		}
		return users
	}
	return nil
}

func BroadcastMessage(payload []byte, users []*User) {
	if users != nil {
		for _, user := range users {
			user.Send <- payload
		}
	}
}
