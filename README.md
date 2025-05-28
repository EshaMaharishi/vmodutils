

<p align="center">
  <a href="https://pkg.go.dev/github.com/erh/vmodutils"><img src="https://pkg.go.dev/badge/github.com/erh/vmodutils" alt="PkgGoDev"></a>
</a>
</p>

## pc crop camera
```
{
  "src" : "<cam>",
  "src_frame" : <optional>, // src point cloud will be converted to world from this, if not specified assume it is world
  "min" : { "X" : 0, "Y" : 0, "Z" : 0}, // specified in world frame
  "min" : { "X" : 9, "Y" : 9, "Z" : 9}  // specified in world frame
}
  
```

## arm position saver
```
{
    "arm" : "<name of arm>", // required
    "motion" : "<name of motion service>", // optional - if used uses post not joines
    "joints" : [ ], // set automatically
    "point" : < ... >,
    "orientation" : < ... >
}
```



## pc multiple arm poses
```
{
 "src" : "<name of camera>",
 "positions" : [ <arm-position-saver>, ... ]
 }
```

## obstacle
Configure this with a frame and you can have obstacles on your robot without having to hard code.
```
{
 "geometries" : [ { "type" : "box", "x" : 100, "y": 100, "z" : 100 } ]
 "geometries" : [ { "type" : "sphere", "r" : 100 } ]

}
```
