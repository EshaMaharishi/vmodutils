package touch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang/geo/r3"

	"go.viam.com/rdk/components/camera"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/pointcloud"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/rimage"
	"go.viam.com/rdk/robot"
	"go.viam.com/rdk/spatialmath"

	"github.com/erh/vmodutils"
)

var CropCameraModel = vmodutils.NamespaceFamily.WithModel("pc-crop-camera")

func init() {
	resource.RegisterComponent(
		camera.API,
		CropCameraModel,
		resource.Registration[camera.Camera, *CropCameraConfig]{
			Constructor: newCropCamera,
		})
}

type CropCameraConfig struct {
	Src      string
	SrcFrame string `json:"src_frame"`
	Min      r3.Vector
	Max      r3.Vector
}

func (ccc *CropCameraConfig) Validate(path string) ([]string, []string, error) {
	if ccc.Src == "" {
		return nil, nil, fmt.Errorf("need a src camera")
	}
	return []string{ccc.Src}, nil, nil
}

func newCropCamera(ctx context.Context, deps resource.Dependencies, config resource.Config, logger logging.Logger) (camera.Camera, error) {
	newConf, err := resource.NativeConfig[*CropCameraConfig](config)
	if err != nil {
		return nil, err
	}

	cc := &cropCamera{
		name:   config.ResourceName(),
		cfg:    newConf,
		logger: logger,
	}

	cc.src, err = camera.FromDependencies(deps, newConf.Src)
	if err != nil {
		return nil, err
	}

	cc.client, err = vmodutils.ConnectToMachineFromEnv(ctx, logger)
	if err != nil {
		return nil, err
	}

	return cc, nil
}

type cropCamera struct {
	resource.AlwaysRebuild

	name   resource.Name
	cfg    *CropCameraConfig
	logger logging.Logger

	src    camera.Camera
	client robot.Robot

	lock               sync.Mutex
	active             bool
	lastPointCloud     pointcloud.PointCloud
	lastPointCloudTime time.Time
	lastPointCloudErr  error
}

func (cc *cropCamera) Name() resource.Name {
	return cc.name
}

func (cc *cropCamera) Image(ctx context.Context, mimeType string, extra map[string]interface{}) ([]byte, camera.ImageMetadata, error) {
	pc, err := cc.NextPointCloud(ctx)
	if err != nil {
		return nil, camera.ImageMetadata{}, err
	}
	img := PCToImage(pc)

	data, err := rimage.EncodeImage(ctx, img, mimeType)
	if err != nil {
		return nil, camera.ImageMetadata{}, err
	}

	return data, camera.ImageMetadata{mimeType}, err
}

func (cc *cropCamera) Images(ctx context.Context, filterSourceNames []string, extra map[string]interface{}) ([]camera.NamedImage, resource.ResponseMetadata, error) {
	pc, err := cc.NextPointCloud(ctx)
	if err != nil {
		return nil, resource.ResponseMetadata{}, err
	}
	start := time.Now()
	img := PCToImage(pc)
	elapsed := time.Since(start)
	if elapsed > (time.Millisecond * 100) {
		cc.logger.Infof("PCToImage took %v", elapsed)
	}
	ni, err := camera.NamedImageFromImage(img, "cropped", "image/png")
	if err != nil {
		return nil, resource.ResponseMetadata{}, err
	}
	return []camera.NamedImage{ni}, resource.ResponseMetadata{time.Now()}, nil
}

func (cc *cropCamera) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}

func (cc *cropCamera) NextPointCloud(ctx context.Context) (pointcloud.PointCloud, error) {

	start := time.Now()
	cc.lock.Lock()
	if cc.active {
		cc.lock.Unlock()
		return cc.waitForPointCloudAfter(ctx, start)
	}

	cc.active = true
	cc.lock.Unlock()

	pc, err := cc.doNextPointCloud(ctx)

	cc.lock.Lock()
	cc.lastPointCloud = pc
	cc.lastPointCloudErr = err
	cc.lastPointCloudTime = time.Now()
	cc.lock.Unlock()

	return pc, err
}

func (cc *cropCamera) waitForPointCloudAfter(ctx context.Context, when time.Time) (pointcloud.PointCloud, error) {
	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if time.Since(when) > time.Minute {
			return nil, fmt.Errorf("waitForPointCloudAfter timed out after %v", time.Since(when))
		}

		cc.lock.Lock()
		if cc.lastPointCloudTime.After(when) {
			pc := cc.lastPointCloud
			err := cc.lastPointCloudErr
			cc.lock.Unlock()
			return pc, err
		}
		cc.lock.Unlock()

		time.Sleep(time.Millisecond * 50)
	}
}

func (cc *cropCamera) doNextPointCloud(ctx context.Context) (pointcloud.PointCloud, error) {
	start := time.Now()

	pc, err := cc.src.NextPointCloud(ctx)
	if err != nil {
		return nil, err
	}

	timeA := time.Since(start)

	srcFrame := cc.cfg.Src
	if cc.cfg.SrcFrame != "" {
		srcFrame = cc.cfg.SrcFrame
	}

	pc, err = cc.client.TransformPointCloud(ctx, pc, srcFrame, "world")
	if err != nil {
		return nil, err
	}

	timeB := time.Since(start)

	pc = PCCrop(pc, cc.cfg.Min, cc.cfg.Max)
	timeC := time.Since(start)

	if timeC > (time.Millisecond * 250) {
		cc.logger.Infof("cropCamera::NextPointCloud timeA: %v timeB: %v timeC: %v", timeA, timeB, timeC)
	}

	return pc, nil

}

func (cc *cropCamera) Properties(ctx context.Context) (camera.Properties, error) {
	return camera.Properties{
		SupportsPCD: true,
	}, nil
}

func (cc *cropCamera) Close(ctx context.Context) error {
	return cc.client.Close(ctx)
}

func (cc *cropCamera) Geometries(ctx context.Context, _ map[string]interface{}) ([]spatialmath.Geometry, error) {
	return nil, nil
}
