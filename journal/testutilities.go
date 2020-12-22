package journal

func newTestAccount() *Account {
	root := NewAccount(RootID)

	/*
		Assets
			Current
			Savings
				ISA
		Expenses
			Life
				Groceries
			Fun
				Hobbies
				Dining Out
		Income
			Work
	*/

	componentsAa := []string{"Assets", "Current"}
	componentsAb := []string{"Assets", "Savings", "ISA"}
	componentsEc := []string{"Expenses", "Life", "Groceries"}
	componentsEa := []string{"Expenses", "Fun", "Hobbies"}
	componentsEb := []string{"Expenses", "Fun", "Dining Out"}
	componentsIa := []string{"Income", "Work"}

	aa := root.FindOrCreateAccount(componentsAa)
	ab := root.FindOrCreateAccount(componentsAb)
	ea := root.FindOrCreateAccount(componentsEa)
	eb := root.FindOrCreateAccount(componentsEb)
	ec := root.FindOrCreateAccount(componentsEc)
	ia := root.FindOrCreateAccount(componentsIa)

	aa.Path = aa.CreatePath()
	ab.Path = ab.CreatePath()
	ea.Path = ea.CreatePath()
	eb.Path = eb.CreatePath()
	ec.Path = ec.CreatePath()
	ia.Path = ia.CreatePath()

	return root
}
