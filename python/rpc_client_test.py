import pennant_pb2_grpc as client
import pennant_pb2 as docs
import grpc


def doclient():
    channel = grpc.insecure_channel('localhost:5432')
    stub = client.PennantStub(channel)
    response = stub.GetFlagValue(docs.FlagRequest(Name='my_flag',
                                                  Strings={
                                                      "user_username": "az",
                                                      "hello": "jello"},
                                                  Numbers={"user_id": 11}))
    print("flag client received: %s / %s" % (response.Status,
                                             response.Enabled))
    response = stub.GetFlagValue(docs.FlagRequest(Name='bogus_flag',
                                                  Strings={"hello": "jello"},
                                                  Numbers={"user_id": 11}))
    print("flag client received: %s / %s" % (response.Status,
                                             response.Enabled))


if __name__ == "__main__":
    doclient()
