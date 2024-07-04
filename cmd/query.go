package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:       "query",
	Short:     "Query the service for data",
	RunE:      query,
	Args:      cobra.MatchAll(cobra.MinimumNArgs(1), cobra.OnlyValidArgs),
	ValidArgs: []string{"add", "subtract"},
}

func init() {
	clientCmd.AddCommand(queryCmd)
	queryCmd.PersistentFlags().Int("a", 0, "value for A")
	viper.BindPFlag("a", queryCmd.PersistentFlags().Lookup("a"))
	queryCmd.PersistentFlags().Int("b", 0, "value for B")
	viper.BindPFlag("b", queryCmd.PersistentFlags().Lookup("b"))
}

func query(cmd *cobra.Command, args []string) error {
	//nc, err := newNatsConnection("piggybank-client")
	//if err != nil {
	//	return err
	//}
	//defer nc.Close()

	//req := service.MathRequest{
	//	A: viper.GetInt("a"),
	//	B: viper.GetInt("b"),
	//}

	//if args[0] == "add" {
	//	mr, err := add(req, nc)
	//	if err != nil {
	//		return err
	//	}

	//	fmt.Println(mr.Result)
	//}

	//if args[0] == "subtract" {
	//	mr, err := subtract(req, nc)
	//	if err != nil {
	//		return err
	//	}

	//	fmt.Println(mr.Result)
	//}

	return nil
}

//func add(req service.MathRequest, nc *nats.Conn) (service.MathResponse, error) {
//	var mr service.MathResponse
//	subject := fmt.Sprintf("prime.services.piggybank.%s.math.add.get", ksuid.New().String())
//
//	data, err := json.Marshal(req)
//	if err != nil {
//		return mr, err
//	}
//
//	resp, err := nc.Request(subject, data, time.Duration(1*time.Second))
//	if err != nil {
//		return mr, err
//	}
//
//	if err := json.Unmarshal(resp.Data, &mr); err != nil {
//		return mr, err
//	}
//
//	return mr, nil
//}
//
//func subtract(req service.MathRequest, nc *nats.Conn) (service.MathResponse, error) {
//	var mr service.MathResponse
//	subject := fmt.Sprintf("prime.services.piggybank.%s.math.subtract.get", ksuid.New().String())
//
//	data, err := json.Marshal(req)
//	if err != nil {
//		return mr, err
//	}
//
//	resp, err := nc.Request(subject, data, time.Duration(1*time.Second))
//	if err != nil {
//		return mr, err
//	}
//
//	if err := json.Unmarshal(resp.Data, &mr); err != nil {
//		return mr, err
//	}
//
//	return mr, nil
//}
