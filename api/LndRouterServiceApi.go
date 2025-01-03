package api

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/wallet/service/apiConnect"
	"io"
)

// SendPaymentV2
//
//	@Description: SendPaymentV2 attempts to route a payment described by the passed PaymentRequest to the final destination.
//	The call returns a stream of payment updates. When using this RPC, make sure to set a fee limit, as the default routing fee limit is 0 sats.
//	Without a non-zero fee limit only routes without fees will be attempted which often fails with FAILURE_REASON_NO_ROUTE.
//	@return string
func SendPaymentV2(invoice string, feelimit int64) string {
	conn, clearUp, err := apiConnect.GetConnection("lnd", false)
	if err != nil {
		fmt.Printf("%s did not connect: %v\n", GetTimeNow(), err)
	}
	defer clearUp()
	client := routerrpc.NewRouterClient(conn)
	request := &routerrpc.SendPaymentRequest{
		PaymentRequest: invoice,
		FeeLimitSat:    feelimit,
		TimeoutSeconds: 60,
	}
	stream, err := client.SendPaymentV2(context.Background(), request)
	if err != nil {
		fmt.Printf("%s routerrpc SendPaymentV2 :%v\n", GetTimeNow(), err)
		return "false"
	}
	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				fmt.Printf("%s err == io.EOF, err: %v\n", GetTimeNow(), err)
				return "false"
			}
			fmt.Printf("%s stream Recv err: %v\n", GetTimeNow(), err)
			return "false"
		}
		fmt.Printf("%s %v\n", GetTimeNow(), response)
		return response.PaymentHash
	}
}

// TrackPaymentV2
//
//	@Description: TrackPaymentV2 returns an update stream for the payment identified by the payment hash.
//	@return string
func TrackPaymentV2(payhash string) string {
	conn, clearUp, err := apiConnect.GetConnection("lnd", false)
	if err != nil {
		fmt.Printf("%s did not connect: %v\n", GetTimeNow(), err)
	}
	defer clearUp()
	client := routerrpc.NewRouterClient(conn)
	_payhashByteSlice, _ := hex.DecodeString(payhash)
	request := &routerrpc.TrackPaymentRequest{
		PaymentHash: _payhashByteSlice,
	}
	stream, err := client.TrackPaymentV2(context.Background(), request)

	if err != nil {
		fmt.Printf("%s client.SendPaymentV2 :%v\n", GetTimeNow(), err)
		return MakeJsonErrorResult(TrackPaymentV2Err, err.Error(), nil)
	}
	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				fmt.Printf("%s err == io.EOF, err: %v\n", GetTimeNow(), err)
				return MakeJsonErrorResult(streamRecvInfoErr, err.Error(), nil)
			}
			fmt.Printf("%s stream Recv err: %v\n", GetTimeNow(), err)
			return MakeJsonErrorResult(streamRecvErr, err.Error(), nil)
		}
		fmt.Printf("%s %v\n", GetTimeNow(), response)
		status := response.Status.String()
		return MakeJsonErrorResult(SUCCESS, "", status)
	}
}

// SendToRouteV2
//
//	@Description:SendToRouteV2 attempts to make a payment via the specified route.
//	This method differs from SendPayment in that it allows users to specify a full route manually.
//	This can be used for things like rebalancing, and atomic swaps.
//	@param route
//	skipped function SendToRouteV2 with unsupported parameter or return types
func SendToRouteV2(payhash []byte, route *lnrpc.Route) {
	conn, clearUp, err := apiConnect.GetConnection("lnd", false)
	if err != nil {
		fmt.Printf("%s did not connect: %v\n", GetTimeNow(), err)
	}
	defer clearUp()
	client := routerrpc.NewRouterClient(conn)
	request := &routerrpc.SendToRouteRequest{
		PaymentHash: payhash,
		Route:       route,
	}
	response, err := client.SendToRouteV2(context.Background(), request)
	if err != nil {
		fmt.Printf("%s routerrpc SendToRouteV2 :%v\n", GetTimeNow(), err)
	}
	fmt.Printf("%s %v\n", GetTimeNow(), response)
}

// EstimateRouteFee
//
//	@Description: EstimateRouteFee allows callers to obtain a lower bound w.r.t how much it may cost to send an HTLC to the target end destination.
//	@return string
func EstimateRouteFee(dest string, amtsat int64) string {
	conn, clearUp, err := apiConnect.GetConnection("lnd", false)
	if err != nil {
		fmt.Printf("%s did not connect: %v\n", GetTimeNow(), err)
	}
	defer clearUp()
	client := routerrpc.NewRouterClient(conn)

	bDest, _ := hex.DecodeString(dest)
	request := &routerrpc.RouteFeeRequest{
		Dest:   bDest,
		AmtSat: amtsat,
	}
	response, err := client.EstimateRouteFee(context.Background(), request)
	if err != nil {
		fmt.Printf("%s routerrpc EstimateRouteFee :%v\n", GetTimeNow(), err)
	}
	fmt.Printf("%s  %v\n", GetTimeNow(), response.RoutingFeeMsat)
	return response.String()
}
