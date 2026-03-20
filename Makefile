gen:
	$(MAKE) -C api oapi
	pnpm --prefix dashboard oapi