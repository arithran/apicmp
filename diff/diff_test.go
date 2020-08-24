package diff

/*
func TestReadCSV(t *testing.T) {
	type args struct {
		csvFile string
	}
	tests := []struct {
		name    string
		args    args
		want    []Event
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				csvFile: "./testdata/test1.csv",
			},
			want: []Event{
				{"999", "gOBxKVbwnZ794xXP6nTbXFz0HcbSfxQD", "/video/2387e4d6a7bede9342150d9afbd0d20f"},
				{"999", "727874d308135908a410a4fe9773e648", "/video/22ac00c58d557bd40b3de684f22578de"},
				{"636", "gOBxKVbwnZ794xXP6nTbXFz0HcbSfxQD", "/video/3e3a3ecbf14f85db2c74a3b79452f3f1"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadCSV(tt.args.csvFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadCSV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadCSV() = %v, want %v", got, tt.want)
			}
		})
	}
}
*/
