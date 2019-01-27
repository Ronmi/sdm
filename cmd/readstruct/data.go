package readstruct

// Info records basic info about a struct and its fields.
//
// For a struct like this:
//
//     package wtf
//     type S struct {
//         A int `key:"val"`
//         b string
//         *X
//         Y
//         C struct{ X int }
//     }
//
// Equivilent Info struct will be:
//
//     Info {
//         Package: "wtf",
//         Name: "S",
//         Fields: []Field{
//             Field{
//                 Name: "A",
//                 Exported: true,
//                 Tags: "`key:\"val\"`",
//                 RawType: "int",
//             },
//             Field{
//                 Name: "b",
//                 Exported: false,
//                 RawType: "string",
//             },
//             Field{
//                 Name: "X",
//                 Exported: true,
//                 RawType: "*X",
//             },
//             Field{
//                 Name: "Y",
//                 Exported: true,
//                 RawType: "Y",
//             },
//             Field{
//                 Name: "C",
//                 Exported: true,
//                 RawType: "",
//             },
//         },
//     }
type Info struct {
	Package string
	Name    string
	Fields  []Field
}

// Field records basic info about a field
//
// For a struct like this:
//
//     type S struct {
//         A int `key:"val"`
//         b string
//         *X
//         Y
//         C struct{ X int }
//     }
//
//  RawType is set to empty string for field "C".
type Field struct {
	Name     string
	Exported bool
	Tags     string
	RawType  string
}
