package httpx

// ClampList normalizes page and perPage for list endpoints,
// ensuring page >= 1 and clamping perPage between 1 and max.
// If perPage <= 0, it uses defaultPerPage.
func ClampList(page, perPage, defaultPerPage, maxPerPage int) (int, int) {
	if page < 1 {
		page = 1
	}
	if perPage <= 0 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}
	return page, perPage
}
