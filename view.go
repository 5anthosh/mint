package mint

//ViewContext #
type ViewContext struct {
	path       string
	methods    []string
	handlers   []Handler
	compressed bool
}

//Views #
type Views []ViewContext

//Path #
func (vc *ViewContext) Path(path string) *ViewContext {
	vc.path = path
	return vc
}

//Methods #
func (vc *ViewContext) Methods(methods ...string) *ViewContext {
	vc.methods = append(vc.methods, methods...)
	return vc
}

//Handlers #
func (vc *ViewContext) Handlers(handlers ...Handler) *ViewContext {
	vc.handlers = append(vc.handlers, handlers...)
	return vc
}

//Compressed #
func (vc *ViewContext) Compressed(compressed bool) *ViewContext {
	vc.compressed = compressed
	return vc
}

//Set #
func (vc *ViewContext) Set(vc1 *ViewContext) {
	*vc1 = *vc

}

//NewViewContext #
func NewViewContext() *ViewContext {
	vc := new(ViewContext)
	return vc
}
