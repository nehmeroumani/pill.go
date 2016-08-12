package auth

import "time"

type userAttempts struct {
	numberOfAttempts            int
	occurrenceTimeOfLastAttempt *time.Time
}

var usersAttempts map[int]userAttempts = map[int]userAttempts{}

const maxAttemptsNumber int = 6
const attemptsDeltaTime float64 = 30

func ResetUserAttempts(userId int) {
	if userId != 0 {
		delete(usersAttempts, userId)
	}
}

func AddUserAttempt(userId int) {
	if userId != 0 {
		if _, ok := usersAttempts[userId]; !ok {
			usersAttempts[userId] = userAttempts{}
		}
		uAttempts := usersAttempts[userId]
		now := time.Now()
		var deltaTime time.Duration
		if uAttempts.occurrenceTimeOfLastAttempt != nil {
			deltaTime = now.Sub(*uAttempts.occurrenceTimeOfLastAttempt)
		} else {
			deltaTime = time.Duration(attemptsDeltaTime + 10)
		}
		if deltaTime.Seconds() <= attemptsDeltaTime {
			uAttempts.numberOfAttempts++
			uAttempts.occurrenceTimeOfLastAttempt = &now
		} else {
			uAttempts.numberOfAttempts = 1
			uAttempts.occurrenceTimeOfLastAttempt = &now
		}
		usersAttempts[userId] = uAttempts
	}
}

func IsUserOverpassMaxAttemptsNumber(userId int) bool {
	if userId != 0 {
		if uAttempts, ok := usersAttempts[userId]; ok {
			if uAttempts.numberOfAttempts >= maxAttemptsNumber {
				return true
			}
		}
	}
	return false
}
