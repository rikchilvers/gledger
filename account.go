package main

import "fmt"

type account struct {
	name         string
	quantity     int64
	parent       *account
	children     map[string]*account
	postings     []*posting
	transactions []*transaction
}

func (a account) String() string {
	n := a.asString(0)
	i := 0
	for _, c := range a.children {
		n = fmt.Sprintf("%s\n%s", n, c.asString(i+1))
		i++
	}
	return fmt.Sprintf("%s\n", n)
}

func (a account) asString(level int) string {
	n := fmt.Sprintf(a.name)
	for i := 0; i < level; i++ {
		n = fmt.Sprintf("\t%s", n)
	}
	return n
}

func newAccount(name string) *account {
	return &account{
		name:         name,
		quantity:     0,
		parent:       nil,
		children:     make(map[string]*account),
		postings:     make([]*posting, 0, 2048),
		transactions: make([]*transaction, 0, 1024),
	}
}

// Returns the deepest account found and any remaining components
func (a account) findChildAndDescend(components []string) (*account, []string) {
	fmt.Printf(">>>\nlooking for %s on %s\n\n", components[0], a.name)
	if account, found := a.children[components[0]]; found {
		if len(components) > 1 {
			return account.findChildAndDescend(components[1:])
		}
		// There are no remaining components so we found the deepest child
		return account, nil
	}
	// There are remaining components so return those and the deepest parent
	return &a, components
}

func (a account) findChild(name string) *account {
	if account, found := a.children[name]; found {
		return account
	}
	return nil
}

func (a *account) addChild(account *account) {
	if _, found := a.children[account.name]; found {
		return
	}

	a.children[account.name] = account
}
