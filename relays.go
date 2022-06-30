package mullsox

import "context"

func GetRelays(ctx context.Context) (ret chan MullvadServer) {
	ret = make(chan MullvadServer)
	
}
