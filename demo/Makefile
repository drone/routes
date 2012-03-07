include $(GOROOT)/src/Make.inc

all:
	$(GC) demo.go
	$(LD) demo.[68]

clean:
	rm -rf *.[68] $(ALL)
	rm -rf *.out $(ALL)

run:
	make all
	./[68].out
