package manager

// pageCount returns the total page count advertised by a list envelope's pagination block.
//
// The manager spec models Pagination.pages as optional, so oapi-codegen generates it as a *int.
// A missing value is treated as a single page: auto-paginating iterators compare their local page
// counter against this and stop once page >= pageCount, so returning 1 makes an envelope without a
// page count terminate after the first page rather than looping forever.
func pageCount(pages *int) int {
	if pages == nil {
		return 1
	}
	return *pages
}
