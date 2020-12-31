package expenses

// Balance shows how much the user owes someone or how much other users in a group owe current user
// Positive number means someone owes the user
// Negative - user owes that person
type Balance map[uint]float32 // userID - amount
