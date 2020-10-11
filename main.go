package main

func main() {
	// parser := NewScannerParser()
	parser := NewReaderParser()
	parser.Parse("test.journal")
}
