package diary

//Overview store of the basic diary information
type Overview struct {
	NumRecipes int
}

//GetOverview - get an overview/dashboard
func GetOverview() (Overview Overview) {

	Overview.NumRecipes = 10 //stub
	return

}
