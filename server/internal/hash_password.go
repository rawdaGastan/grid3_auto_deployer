// Package internal for internal details
package internal

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

/*
I don't know why you are using bcrypt over something more modern say sha256. But in either cases
the SALT shouldn't be passed to the hash function NOR it should be chosen or set in the config at all
the salt should be totally randomly generated before computing the hash. hence it will be unique per password

The idea behind the salt to make it impossible to do do a hashtable attack here is how it works and how it should be used:
- If we have passowrd x, and hash(x) = y
- it is impossible to compute x back from y, in other words fn(y) = x does not exist.
- now, someone can create a huge table of all common passwords like `password`, `123456`, etc.. then run a hash function
  on this huge list, and create a hash table where `k = hash(x), v=x`.
- this mean that if I now have the hash `y`, and IF y is one of the easy common password, i can simply do look up `y` in the
  hash table, and then get `x`` since all hashes where pre computed by someone.

Now, to avoid this we use salt: and this is how salt work (and why this function and the way you configure salt is totally wrong)
- I generate salt=random() of fixed length say 10 bytes (every time)
- I then compute y = hash(salt + x)
- Then i create the hashed password (the one i actually store) as `hashed = salt + y` (this can be then hex encoded or base64 it does not matter)
- Now it's IMPOSSIBLE for an attacker to use a hash table attach on the stored `hashed` because even if password is `123456` it will never be found
  on the hash table, right ?
- Now you are wondering how then i can validate the password against hashed, it's very simple
- I can have validate function that looks like this validate(hashed, password)
- Then, it can do this `(salt, y) = (hashed[:10], hashed[10:])`
- Then simply do `return y == hash(salt + x)`

Also notice that also means if more than one user used the same password, their stored hashes will look completely different it means even
if one password is cracked, the other users using the same password won't be affected.
*/

// HashAndSaltPassword hashes password of user
func HashAndSaltPassword(password string, salt string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(salt+password), bcrypt.MinCost)
	if err != nil {
		return "", fmt.Errorf("could not hash password %w", err)
	}
	return string(hashedPassword), nil
}

// VerifyPassword checks if given password is same as hashed one
func VerifyPassword(hashedPassword string, password string, salt string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(salt+password))
	return err == nil
}
