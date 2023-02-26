package utilities

import (
    "fmt"
    "math/rand"
)

var firstNames = []string{"Alice", "Bob", "Charlie", "David", "Eve", "Frank", "Grace", "Heidi", "Ivan", "Julia", "Kevin", "Linda", "Mallory", "Nancy", "Oscar", "Peggy", "Quentin", "Randy", "Sybil", "Trent", "Ursula", "Victor", "Wendy", "Xander", "Yvonne", "Zelda"}

var lastNames = []string{"Smith", "Johnson", "Brown", "Davis", "Wilson", "Kim", "Schmidt", "Petrov", "Rodriguez", "Garcia", "Gonzalez", "Martinez", "Hernandez", "Lopez", "Perez", "Jackson", "Taylor", "Lee", "Nguyen", "Chen", "Wang", "Singh", "Kim", "Gupta", "Kumar"}

func GetRandomName() string {
    firstName := firstNames[rand.Intn(len(firstNames))]
    lastName := lastNames[rand.Intn(len(lastNames))]
    return fmt.Sprintf("%s %s", firstName, lastName)
}

