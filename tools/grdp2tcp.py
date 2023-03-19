import grpc
import argparse

from clientpb import client_pb2_grpc
from commonpb import common_pb2
empty = common_pb2.Empty()

def list_endpoints(args):
    with grpc.insecure_channel(args.control) as channel:
        stub = client_pb2_grpc.ClientRpcStub(channel)

        endpoints = stub.List(empty).endpoints
        print("\n".join(endpoints))

def add_socks(args):
    host,port = args.listen.split(':')
    addr = common_pb2.Addr()
    addr.ip, addr.port = host, int(port)

    with grpc.insecure_channel(args.control) as channel:
        stub = client_pb2_grpc.ClientRpcStub(channel)

        try:
            stub.SocksStart(addr)
        except grpc.RpcError as rpc_error:
            print(rpc_error.details())

def rm_socks(args):
    host,port = args.listen.split(':')
    addr = common_pb2.Addr()
    addr.ip, addr.port = host, int(port)

    with grpc.insecure_channel(args.control) as channel:
        stub = client_pb2_grpc.ClientRpcStub(channel)

        try:
            stub.SocksStop(addr)
        except grpc.RpcError as rpc_error:
            print(rpc_error.details())

def add_reverse(args):
    host,port = args.listen.split(':')
    remote_host,remote_port = args.remote.split(':')
    addrPack = common_pb2.AddrPack()
    addrPack.local.ip, addrPack.local.port = host, int(port)
    addrPack.remote.ip, addrPack.remote.port = remote_host, int(remote_port)

    with grpc.insecure_channel(args.control) as channel:
        stub = client_pb2_grpc.ClientRpcStub(channel)

        try:
            stub.ReverseStart(addrPack)
        except grpc.RpcError as rpc_error:
            print(rpc_error.details())

def rm_reverse(args):
    host,port = args.listen.split(':')
    remote_host,remote_port = args.remote.split(':')
    addrPack = common_pb2.AddrPack()
    addrPack.local.ip, addrPack.local.port = host, int(port)
    addrPack.remote.ip, addrPack.remote.port = remote_host, int(remote_port)

    with grpc.insecure_channel(args.control) as channel:
        stub = client_pb2_grpc.ClientRpcStub(channel)

        try:
            stub.ReverseStop(addrPack.remote)
        except grpc.RpcError as rpc_error:
            print(rpc_error.details())


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        prog = "grdp2tcp.py"
    )

    parser.add_argument("-c", '--control', metavar = "127.0.0.1:8337", default = "127.0.0.1:8337", help = "Control server address")


    main_sub = parser.add_subparsers(title="Commands")

    socks_parser = main_sub.add_parser("socks")
    socks_sub = socks_parser.add_subparsers()
    socks_listen_addr = argparse.ArgumentParser(add_help = False)
    socks_listen_addr.add_argument("-l", "--listen", metavar="127.0.0.1:1080", default = "127.0.0.1:1080", help = "Listen address")
    add_parser = socks_sub.add_parser("add", parents=[socks_listen_addr])
    add_parser.set_defaults(func=add_socks)
    rm_parser = socks_sub.add_parser("rm", parents=[socks_listen_addr])
    rm_parser.set_defaults(func=rm_socks)

    reverse_parser = main_sub.add_parser("reverse")
    reverse_sub = reverse_parser.add_subparsers()
    remote_listen_addr = argparse.ArgumentParser(add_help = False)
    remote_listen_addr.add_argument("-l", "--listen", metavar="127.0.0.1:8445", help = "Local connect address", required = True)
    remote_addr = argparse.ArgumentParser(add_help = False)
    remote_addr.add_argument("-r", "--remote", metavar="127.0.0.1:8445", help = "Remote listen address", required = True)
    r_add_parser = reverse_sub.add_parser("add", parents=[remote_listen_addr, remote_addr])
    r_add_parser.set_defaults(func=add_reverse)
    r_rm_parser = reverse_sub.add_parser("rm", parents=[remote_listen_addr, remote_addr])
    r_rm_parser.set_defaults(func=rm_reverse)

    list_parser = main_sub.add_parser("list")
    list_parser.set_defaults(func=list_endpoints)
    args = parser.parse_args()

    args.func(args)
