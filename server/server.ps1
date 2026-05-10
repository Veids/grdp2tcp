param(
    [string]$ChannelName = "rdp2tcp",
    [string]$SftpServerPath = "",
    [switch]$NoSftp,
    [switch]$CompileOnly
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$source = @'
using System;
using System.Collections.Concurrent;
using System.ComponentModel;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Net;
using System.Net.Sockets;
using System.Runtime.InteropServices;
using System.Text;
using System.Threading;
using System.Threading.Tasks;

namespace Grdp2Tcp.PowerShellServer
{
    internal static class Log
    {
        public static void Info(string message)
        {
            Console.Error.WriteLine(DateTime.Now.ToString("s") + " " + message);
        }
    }

    internal sealed class VirtualChannelStream : Stream
    {
        private const int WTS_CURRENT_SESSION = -1;
        private const int WTSVirtualFileHandle = 1;
        private const int ERROR_HANDLE_EOF = 38;
        private const int ERROR_BROKEN_PIPE = 109;
        private const int ERROR_MORE_DATA = 234;
        private const int ERROR_IO_PENDING = 997;
        private const uint WAIT_OBJECT_0 = 0;
        private const uint INFINITE = 0xFFFFFFFF;
        private const int ChannelReadBufferSize = 64 * 1024;

        private readonly IntPtr channelHandle;
        private readonly IntPtr fileHandle;
        private readonly object readLock = new object();
        private readonly object writeLock = new object();
        private byte[] pendingRead;
        private int pendingReadOffset;
        private int pendingReadLength;

        [StructLayout(LayoutKind.Sequential)]
        private struct NativeOverlapped
        {
            public IntPtr Internal;
            public IntPtr InternalHigh;
            public uint Offset;
            public uint OffsetHigh;
            public IntPtr EventHandle;
        }

        [DllImport("wtsapi32.dll", SetLastError = true, CharSet = CharSet.Ansi)]
        private static extern IntPtr WTSVirtualChannelOpenEx(int sessionId, string virtualName, int flags);

        [DllImport("wtsapi32.dll", SetLastError = true)]
        private static extern bool WTSVirtualChannelQuery(
            IntPtr channelHandle,
            int wtsVirtualClass,
            out IntPtr buffer,
            out int bytesReturned);

        [DllImport("wtsapi32.dll", SetLastError = true)]
        private static extern void WTSFreeMemory(IntPtr memory);

        [DllImport("wtsapi32.dll", SetLastError = true)]
        private static extern bool WTSVirtualChannelClose(IntPtr channelHandle);

        [DllImport("kernel32.dll", SetLastError = true)]
        private static extern IntPtr CreateEvent(IntPtr eventAttributes, bool manualReset, bool initialState, string name);

        [DllImport("kernel32.dll", SetLastError = true)]
        private static extern bool CloseHandle(IntPtr handle);

        [DllImport("kernel32.dll", SetLastError = true)]
        private static extern uint WaitForSingleObject(IntPtr handle, uint milliseconds);

        [DllImport("kernel32.dll", SetLastError = true)]
        private static extern bool GetOverlappedResult(
            IntPtr handle,
            ref NativeOverlapped overlapped,
            out int bytesTransferred,
            bool wait);

        [DllImport("kernel32.dll", SetLastError = true)]
        private static extern bool ReadFile(
            IntPtr handle,
            byte[] buffer,
            int bytesToRead,
            IntPtr bytesRead,
            ref NativeOverlapped overlapped);

        [DllImport("kernel32.dll", SetLastError = true)]
        private static extern bool WriteFile(
            IntPtr handle,
            byte[] buffer,
            int bytesToWrite,
            IntPtr bytesWritten,
            ref NativeOverlapped overlapped);

        public VirtualChannelStream(string channelName)
        {
            channelHandle = WTSVirtualChannelOpenEx(WTS_CURRENT_SESSION, channelName, 0);
            if (channelHandle == IntPtr.Zero)
            {
                throw new InvalidOperationException("WTSVirtualChannelOpenEx failed: " + Marshal.GetLastWin32Error());
            }

            IntPtr buffer = IntPtr.Zero;
            int bytesReturned = 0;
            if (!WTSVirtualChannelQuery(channelHandle, WTSVirtualFileHandle, out buffer, out bytesReturned))
            {
                int error = Marshal.GetLastWin32Error();
                WTSVirtualChannelClose(channelHandle);
                throw new InvalidOperationException("WTSVirtualChannelQuery failed: " + error);
            }

            try
            {
                fileHandle = Marshal.ReadIntPtr(buffer);
            }
            finally
            {
                if (buffer != IntPtr.Zero)
                {
                    WTSFreeMemory(buffer);
                }
            }
        }

        public override bool CanRead { get { return true; } }
        public override bool CanSeek { get { return false; } }
        public override bool CanWrite { get { return true; } }
        public override long Length { get { throw new NotSupportedException(); } }
        public override long Position
        {
            get { throw new NotSupportedException(); }
            set { throw new NotSupportedException(); }
        }

        public override void Flush()
        {
        }

        public override int Read(byte[] buffer, int offset, int count)
        {
            if (count == 0)
            {
                return 0;
            }

            lock (readLock)
            {
                if (pendingRead != null && pendingReadOffset < pendingReadLength)
                {
                    return CopyPendingRead(buffer, offset, count);
                }

                byte[] channelBuffer = new byte[ChannelReadBufferSize];
                int read = ReadNative(channelBuffer, channelBuffer.Length);
                if (read <= 0)
                {
                    return read;
                }

                pendingRead = channelBuffer;
                pendingReadOffset = 0;
                pendingReadLength = read;
                return CopyPendingRead(buffer, offset, count);
            }
        }

        public override long Seek(long offset, SeekOrigin origin)
        {
            throw new NotSupportedException();
        }

        public override void SetLength(long value)
        {
            throw new NotSupportedException();
        }

        public override void Write(byte[] buffer, int offset, int count)
        {
            lock (writeLock)
            {
                int written = 0;
                while (written < count)
                {
                    int chunk = count - written;
                    byte[] writeBuffer = new byte[chunk];
                    Buffer.BlockCopy(buffer, offset + written, writeBuffer, 0, chunk);
                    int transferred = WriteNative(writeBuffer, chunk);
                    if (transferred <= 0)
                    {
                        throw new IOException("virtual channel write returned 0 bytes");
                    }
                    written += transferred;
                }
            }
        }

        protected override void Dispose(bool disposing)
        {
            if (channelHandle != IntPtr.Zero)
            {
                WTSVirtualChannelClose(channelHandle);
            }
            base.Dispose(disposing);
        }

        private int CopyPendingRead(byte[] buffer, int offset, int count)
        {
            int available = pendingReadLength - pendingReadOffset;
            int copied = Math.Min(count, available);
            Buffer.BlockCopy(pendingRead, pendingReadOffset, buffer, offset, copied);
            pendingReadOffset += copied;
            if (pendingReadOffset >= pendingReadLength)
            {
                pendingRead = null;
                pendingReadOffset = 0;
                pendingReadLength = 0;
            }
            return copied;
        }

        private int ReadNative(byte[] buffer, int count)
        {
            return NativeTransfer(true, buffer, count);
        }

        private int WriteNative(byte[] buffer, int count)
        {
            return NativeTransfer(false, buffer, count);
        }

        private int NativeTransfer(bool read, byte[] buffer, int count)
        {
            IntPtr eventHandle = CreateEvent(IntPtr.Zero, true, false, null);
            if (eventHandle == IntPtr.Zero)
            {
                throw Win32IOException("CreateEvent", Marshal.GetLastWin32Error());
            }

            GCHandle pinnedBuffer = GCHandle.Alloc(buffer, GCHandleType.Pinned);
            try
            {
                NativeOverlapped overlapped = new NativeOverlapped();
                overlapped.EventHandle = eventHandle;

                bool completed = read
                    ? ReadFile(fileHandle, buffer, count, IntPtr.Zero, ref overlapped)
                    : WriteFile(fileHandle, buffer, count, IntPtr.Zero, ref overlapped);
                int error = completed ? 0 : Marshal.GetLastWin32Error();

                if (!completed && error == ERROR_IO_PENDING)
                {
                    uint wait = WaitForSingleObject(eventHandle, INFINITE);
                    if (wait != WAIT_OBJECT_0)
                    {
                        throw Win32IOException("WaitForSingleObject", Marshal.GetLastWin32Error());
                    }
                }
                else if (!completed && error != ERROR_MORE_DATA)
                {
                    if (error == ERROR_BROKEN_PIPE || error == ERROR_HANDLE_EOF)
                    {
                        return 0;
                    }
                    throw Win32IOException(read ? "ReadFile" : "WriteFile", error);
                }

                int bytesTransferred;
                if (GetOverlappedResult(fileHandle, ref overlapped, out bytesTransferred, false))
                {
                    return bytesTransferred;
                }

                int resultError = Marshal.GetLastWin32Error();
                if (read && (resultError == ERROR_MORE_DATA || error == ERROR_MORE_DATA) && bytesTransferred > 0)
                {
                    return bytesTransferred;
                }
                if (resultError == ERROR_BROKEN_PIPE || resultError == ERROR_HANDLE_EOF)
                {
                    return 0;
                }

                throw Win32IOException(read ? "GetOverlappedResult(ReadFile)" : "GetOverlappedResult(WriteFile)", resultError);
            }
            finally
            {
                pinnedBuffer.Free();
                CloseHandle(eventHandle);
            }
        }

        private static IOException Win32IOException(string operation, int error)
        {
            return new IOException(operation + " failed with Windows error " + error + ": " + new Win32Exception(error).Message);
        }
    }

    internal sealed class YamuxSession
    {
        private const byte TypeData = 0;
        private const byte TypeWindowUpdate = 1;
        private const byte TypePing = 2;
        private const byte TypeGoAway = 3;
        private const ushort FlagSyn = 1;
        private const ushort FlagAck = 2;
        private const ushort FlagFin = 4;
        private const ushort FlagRst = 8;
        private const uint InitialWindow = 256 * 1024;
        private const int HeaderSize = 12;
        private const int WriteChunk = 16 * 1024;

        private readonly Stream conn;
        private readonly object writeLock = new object();
        private readonly ConcurrentDictionary<uint, YamuxStream> streams = new ConcurrentDictionary<uint, YamuxStream>();
        private readonly Action<YamuxStream> streamHandler;
        private volatile bool closed;

        public YamuxSession(Stream conn, Action<YamuxStream> streamHandler)
        {
            this.conn = conn;
            this.streamHandler = streamHandler;
        }

        public void Serve()
        {
            while (!closed)
            {
                byte[] header = ReadExact(conn, HeaderSize);
                byte version = header[0];
                if (version != 0)
                {
                    throw new InvalidDataException("unsupported yamux protocol version " + version);
                }

                byte msgType = header[1];
                ushort flags = ReadUInt16BE(header, 2);
                uint streamId = ReadUInt32BE(header, 4);
                uint length = ReadUInt32BE(header, 8);

                if (msgType == TypePing)
                {
                    if ((flags & FlagSyn) != 0)
                    {
                        SendFrame(TypePing, FlagAck, 0, length, null, 0, 0);
                    }
                    continue;
                }

                if (msgType == TypeGoAway)
                {
                    closed = true;
                    return;
                }

                if (msgType != TypeData && msgType != TypeWindowUpdate)
                {
                    throw new InvalidDataException("unsupported yamux message type " + msgType);
                }

                YamuxStream stream = null;
                if ((flags & FlagSyn) != 0)
                {
                    stream = new YamuxStream(this, streamId, InitialWindow);
                    if (!streams.TryAdd(streamId, stream))
                    {
                        SendFrame(TypeWindowUpdate, FlagRst, streamId, 0, null, 0, 0);
                        try { stream.DisposeUnregistered(); } catch { }
                        continue;
                    }
                    SendFrame(TypeWindowUpdate, FlagAck, streamId, 0, null, 0, 0);
                    Task.Factory.StartNew(delegate { streamHandler(stream); }, TaskCreationOptions.LongRunning);
                }
                else
                {
                    streams.TryGetValue(streamId, out stream);
                }

                if (stream == null)
                {
                    if (msgType == TypeData && length > 0)
                    {
                        Drain(conn, length);
                    }
                    continue;
                }

                if (msgType == TypeWindowUpdate)
                {
                    if ((flags & FlagRst) != 0)
                    {
                        stream.RemoteReset();
                        streams.TryRemove(streamId, out stream);
                    }
                    else
                    {
                        stream.AddSendWindow(length);
                        if ((flags & FlagFin) != 0)
                        {
                            stream.RemoteClose();
                        }
                    }
                    continue;
                }

                byte[] payload = length == 0 ? new byte[0] : ReadExact(conn, checked((int)length));
                if (payload.Length > 0)
                {
                    stream.Enqueue(payload);
                    SendFrame(TypeWindowUpdate, 0, streamId, (uint)payload.Length, null, 0, 0);
                }
                if ((flags & FlagFin) != 0)
                {
                    stream.RemoteClose();
                }
            }
        }

        public void WriteData(uint streamId, byte[] buffer, int offset, int count, Func<int, int> reserveWindow)
        {
            int sent = 0;
            while (sent < count)
            {
                int wanted = Math.Min(WriteChunk, count - sent);
                int chunk = reserveWindow(wanted);
                SendFrame(TypeData, 0, streamId, (uint)chunk, buffer, offset + sent, chunk);
                sent += chunk;
            }
        }

        public void CloseStream(YamuxStream stream)
        {
            uint streamId = stream.StreamId;
            SendFrame(TypeWindowUpdate, FlagFin, streamId, 0, null, 0, 0);
            YamuxStream existing;
            if (streams.TryGetValue(streamId, out existing) && ReferenceEquals(existing, stream))
            {
                YamuxStream ignored;
                streams.TryRemove(streamId, out ignored);
            }
        }

        private void SendFrame(byte msgType, ushort flags, uint streamId, uint length, byte[] body, int offset, int count)
        {
            byte[] header = new byte[HeaderSize];
            header[0] = 0;
            header[1] = msgType;
            WriteUInt16BE(header, 2, flags);
            WriteUInt32BE(header, 4, streamId);
            WriteUInt32BE(header, 8, length);

            lock (writeLock)
            {
                conn.Write(header, 0, header.Length);
                if (count > 0)
                {
                    conn.Write(body, offset, count);
                }
                conn.Flush();
            }
        }

        private static void Drain(Stream stream, uint count)
        {
            byte[] buffer = new byte[8192];
            uint remaining = count;
            while (remaining > 0)
            {
                int read = stream.Read(buffer, 0, (int)Math.Min((uint)buffer.Length, remaining));
                if (read <= 0)
                {
                    throw new EndOfStreamException();
                }
                remaining -= (uint)read;
            }
        }

        internal static byte[] ReadExact(Stream stream, int count)
        {
            byte[] buffer = new byte[count];
            int offset = 0;
            while (offset < count)
            {
                int read = stream.Read(buffer, offset, count - offset);
                if (read <= 0)
                {
                    throw new EndOfStreamException();
                }
                offset += read;
            }
            return buffer;
        }

        private static ushort ReadUInt16BE(byte[] buffer, int offset)
        {
            return (ushort)((buffer[offset] << 8) | buffer[offset + 1]);
        }

        private static uint ReadUInt32BE(byte[] buffer, int offset)
        {
            return ((uint)buffer[offset] << 24) |
                   ((uint)buffer[offset + 1] << 16) |
                   ((uint)buffer[offset + 2] << 8) |
                   buffer[offset + 3];
        }

        private static void WriteUInt16BE(byte[] buffer, int offset, ushort value)
        {
            buffer[offset] = (byte)(value >> 8);
            buffer[offset + 1] = (byte)value;
        }

        private static void WriteUInt32BE(byte[] buffer, int offset, uint value)
        {
            buffer[offset] = (byte)(value >> 24);
            buffer[offset + 1] = (byte)(value >> 16);
            buffer[offset + 2] = (byte)(value >> 8);
            buffer[offset + 3] = (byte)value;
        }
    }

    internal sealed class YamuxStream : Stream
    {
        private readonly YamuxSession session;
        private readonly uint id;
        private readonly ConcurrentQueue<byte[]> queue = new ConcurrentQueue<byte[]>();
        private readonly SemaphoreSlim available = new SemaphoreSlim(0);
        private readonly object windowLock = new object();

        private byte[] current;
        private int currentOffset;
        private bool remoteClosed;
        private bool localClosed;
        private bool disposed;
        private bool reset;
        private uint sendWindow;

        public YamuxStream(YamuxSession session, uint id, uint initialWindow)
        {
            this.session = session;
            this.id = id;
            sendWindow = initialWindow;
        }

        internal uint StreamId { get { return id; } }

        public override bool CanRead { get { return true; } }
        public override bool CanSeek { get { return false; } }
        public override bool CanWrite { get { return true; } }
        public override long Length { get { throw new NotSupportedException(); } }
        public override long Position
        {
            get { throw new NotSupportedException(); }
            set { throw new NotSupportedException(); }
        }

        public void Enqueue(byte[] bytes)
        {
            queue.Enqueue(bytes);
            available.Release();
        }

        public void AddSendWindow(uint delta)
        {
            if (delta == 0)
            {
                return;
            }
            lock (windowLock)
            {
                sendWindow += delta;
                Monitor.PulseAll(windowLock);
            }
        }

        public void RemoteClose()
        {
            remoteClosed = true;
            try { available.Release(); } catch { }
        }

        public void RemoteReset()
        {
            lock (windowLock)
            {
                reset = true;
                Monitor.PulseAll(windowLock);
            }
            try { available.Release(); } catch { }
        }

        public override int Read(byte[] buffer, int offset, int count)
        {
            while (true)
            {
                if (reset)
                {
                    throw new IOException("yamux stream reset");
                }

                if (current != null && currentOffset < current.Length)
                {
                    int n = Math.Min(count, current.Length - currentOffset);
                    Buffer.BlockCopy(current, currentOffset, buffer, offset, n);
                    currentOffset += n;
                    if (currentOffset >= current.Length)
                    {
                        current = null;
                        currentOffset = 0;
                    }
                    return n;
                }

                byte[] next;
                if (queue.TryDequeue(out next))
                {
                    current = next;
                    currentOffset = 0;
                    continue;
                }

                if (remoteClosed)
                {
                    return 0;
                }

                try
                {
                    available.Wait();
                }
                catch (ObjectDisposedException)
                {
                    if (localClosed || disposed)
                    {
                        throw new ObjectDisposedException("YamuxStream");
                    }
                    throw;
                }
                if (localClosed || disposed)
                {
                    throw new ObjectDisposedException("YamuxStream");
                }
            }
        }

        public override void Write(byte[] buffer, int offset, int count)
        {
            if (localClosed)
            {
                throw new ObjectDisposedException("YamuxStream");
            }
            session.WriteData(id, buffer, offset, count, ReserveWindow);
        }

        private int ReserveWindow(int wanted)
        {
            lock (windowLock)
            {
                while (true)
                {
                    if (reset)
                    {
                        throw new IOException("yamux stream reset");
                    }
                    if (localClosed || disposed)
                    {
                        throw new ObjectDisposedException("YamuxStream");
                    }
                    if (sendWindow > 0)
                    {
                        int reserved = Math.Min(wanted, (int)Math.Min((uint)(16 * 1024), sendWindow));
                        sendWindow -= (uint)reserved;
                        return reserved;
                    }
                    Monitor.Wait(windowLock, 1000);
                }
            }
        }

        public override void Flush() { }
        public override long Seek(long offset, SeekOrigin origin) { throw new NotSupportedException(); }
        public override void SetLength(long value) { throw new NotSupportedException(); }

        internal void DisposeUnregistered()
        {
            if (disposed)
            {
                return;
            }
            localClosed = true;
            lock (windowLock)
            {
                Monitor.PulseAll(windowLock);
            }
            try { available.Release(); } catch { }
            available.Dispose();
            disposed = true;
            base.Dispose(true);
        }

        protected override void Dispose(bool disposing)
        {
            if (disposed)
            {
                return;
            }
            if (disposing)
            {
                if (!localClosed)
                {
                    localClosed = true;
                    session.CloseStream(this);
                }
                lock (windowLock)
                {
                    Monitor.PulseAll(windowLock);
                }
                try { available.Release(); } catch { }
                available.Dispose();
            }
            disposed = true;
            base.Dispose(disposing);
        }
    }

    internal sealed class Server
    {
        private readonly string sftpServerPath;
        private readonly bool noSftp;

        public Server(string sftpServerPath, bool noSftp)
        {
            this.sftpServerPath = sftpServerPath;
            this.noSftp = noSftp;
        }

        public void HandleStream(YamuxStream stream)
        {
            try
            {
                byte[] kind = YamuxSession.ReadExact(stream, 1);
                switch (kind[0])
                {
                    case 0:
                        Log.Info("control stream accepted");
                        HoldControlStream(stream);
                        break;
                    case 1:
                        Log.Info("SOCKS stream accepted");
                        ServeSocks(stream);
                        break;
                    case 2:
                        Log.Info("forward stream requested, but Go gob address decoding is not implemented in server.ps1");
                        stream.Dispose();
                        break;
                    case 4:
                        Log.Info("SFTP stream accepted");
                        ServeSftp(stream);
                        break;
                    default:
                        Log.Info("invalid stream type " + kind[0]);
                        stream.Dispose();
                        break;
                }
            }
            catch (Exception ex)
            {
                Log.Info("stream failed: " + ex.Message);
                try { stream.Dispose(); } catch { }
            }
        }

        private static void HoldControlStream(Stream stream)
        {
            byte[] buffer = new byte[1024];
            try
            {
                while (stream.Read(buffer, 0, buffer.Length) > 0) { }
            }
            catch { }
            finally
            {
                try { stream.Dispose(); } catch { }
            }
        }

        private void ServeSftp(YamuxStream stream)
        {
            if (noSftp)
            {
                Log.Info("SFTP disabled by -NoSftp");
                stream.Dispose();
                return;
            }

            string path = ResolveSftpServerPath(sftpServerPath);
            if (String.IsNullOrEmpty(path))
            {
                Log.Info("sftp-server.exe was not found; install OpenSSH Server or pass -SftpServerPath");
                stream.Dispose();
                return;
            }

            ProcessStartInfo startInfo = new ProcessStartInfo();
            startInfo.FileName = path;
            startInfo.UseShellExecute = false;
            startInfo.RedirectStandardInput = true;
            startInfo.RedirectStandardOutput = true;
            startInfo.RedirectStandardError = true;
            startInfo.CreateNoWindow = true;

            Process process = null;
            try
            {
                process = Process.Start(startInfo);
                Task pumpIn = Task.Factory.StartNew(delegate { Pump(stream, process.StandardInput.BaseStream); }, TaskCreationOptions.LongRunning);
                Task pumpOut = Task.Factory.StartNew(delegate { Pump(process.StandardOutput.BaseStream, stream); }, TaskCreationOptions.LongRunning);
                Task pumpErr = Task.Factory.StartNew(delegate
                {
                    try
                    {
                        using (StreamReader err = process.StandardError)
                        {
                            string error = err.ReadToEnd();
                            if (!String.IsNullOrEmpty(error))
                            {
                                Log.Info("sftp-server stderr: " + error.Trim());
                            }
                        }
                    }
                    catch { }
                }, TaskCreationOptions.LongRunning);
                Task.WaitAll(pumpIn, pumpOut, pumpErr);
            }
            finally
            {
                if (process != null)
                {
                    try { process.Dispose(); } catch { }
                }
                try { stream.Dispose(); } catch { }
            }
        }

        private static string ResolveSftpServerPath(string configured)
        {
            if (!String.IsNullOrWhiteSpace(configured) && File.Exists(configured))
            {
                return configured;
            }

            string windir = Environment.GetFolderPath(Environment.SpecialFolder.Windows);
            string[] candidates = new string[]
            {
                Path.Combine(windir, "System32", "OpenSSH", "sftp-server.exe"),
                Path.Combine(windir, "Sysnative", "OpenSSH", "sftp-server.exe"),
                @"C:\Program Files\OpenSSH-Win64\sftp-server.exe"
            };

            for (int i = 0; i < candidates.Length; i++)
            {
                if (File.Exists(candidates[i]))
                {
                    return candidates[i];
                }
            }
            return null;
        }

        private static void ServeSocks(YamuxStream stream)
        {
            TcpClient tcp = null;
            try
            {
                byte[] greeting = YamuxSession.ReadExact(stream, 2);
                if (greeting[0] != 5)
                {
                    throw new InvalidDataException("invalid SOCKS version");
                }
                YamuxSession.ReadExact(stream, greeting[1]);
                stream.Write(new byte[] { 5, 0 }, 0, 2);

                byte[] request = YamuxSession.ReadExact(stream, 4);
                if (request[0] != 5)
                {
                    throw new InvalidDataException("invalid SOCKS request version");
                }
                if (request[1] != 1)
                {
                    WriteSocksFailure(stream, 7);
                    return;
                }

                string host;
                byte atyp = request[3];
                if (atyp == 1)
                {
                    host = new IPAddress(YamuxSession.ReadExact(stream, 4)).ToString();
                }
                else if (atyp == 3)
                {
                    int length = YamuxSession.ReadExact(stream, 1)[0];
                    host = Encoding.ASCII.GetString(YamuxSession.ReadExact(stream, length));
                }
                else if (atyp == 4)
                {
                    host = new IPAddress(YamuxSession.ReadExact(stream, 16)).ToString();
                }
                else
                {
                    WriteSocksFailure(stream, 8);
                    return;
                }

                byte[] portBytes = YamuxSession.ReadExact(stream, 2);
                int port = (portBytes[0] << 8) | portBytes[1];

                tcp = new TcpClient();
                tcp.Connect(host, port);
                WriteSocksSuccess(stream, (IPEndPoint)tcp.Client.RemoteEndPoint);
                Log.Info("SOCKS connected to " + host + ":" + port);

                NetworkStream network = tcp.GetStream();
                Task a = Task.Factory.StartNew(delegate { Pump(stream, network); }, TaskCreationOptions.LongRunning);
                Task b = Task.Factory.StartNew(delegate { Pump(network, stream); }, TaskCreationOptions.LongRunning);
                Task.WaitAll(a, b);
            }
            catch (Exception ex)
            {
                Log.Info("SOCKS failed: " + ex.Message);
                try { WriteSocksFailure(stream, 1); } catch { }
            }
            finally
            {
                if (tcp != null)
                {
                    try { tcp.Close(); } catch { }
                }
                try { stream.Dispose(); } catch { }
            }
        }

        private static void WriteSocksFailure(Stream stream, byte code)
        {
            byte[] response = new byte[] { 5, code, 0, 1, 0, 0, 0, 0, 0, 0 };
            stream.Write(response, 0, response.Length);
        }

        private static void WriteSocksSuccess(Stream stream, IPEndPoint bndEndPoint)
        {
            byte[] address = bndEndPoint.Address.GetAddressBytes();
            byte atyp = address.Length == 16 ? (byte)4 : (byte)1;
            byte[] response = new byte[4 + address.Length + 2];
            response[0] = 5;
            response[1] = 0;
            response[2] = 0;
            response[3] = atyp;
            Buffer.BlockCopy(address, 0, response, 4, address.Length);
            response[response.Length - 2] = (byte)(bndEndPoint.Port >> 8);
            response[response.Length - 1] = (byte)bndEndPoint.Port;
            stream.Write(response, 0, response.Length);
        }

        private static void Pump(Stream input, Stream output)
        {
            byte[] buffer = new byte[16 * 1024];
            try
            {
                while (true)
                {
                    int read = input.Read(buffer, 0, buffer.Length);
                    if (read <= 0)
                    {
                        break;
                    }
                    output.Write(buffer, 0, read);
                    output.Flush();
                }
            }
            catch { }
        }
    }

    public static class Program
    {
        public static void CompileOnly()
        {
            Log.Info("server.ps1 helper compiled");
        }

        public static void Run(string channelName, string sftpServerPath, bool noSftp)
        {
            using (VirtualChannelStream channel = new VirtualChannelStream(channelName))
            {
                Log.Info("VC channel connected");
                Challenge(channel);
                Log.Info("client challenge completed");

                Server server = new Server(sftpServerPath, noSftp);
                YamuxSession session = new YamuxSession(channel, server.HandleStream);
                session.Serve();
            }
        }

        private static void Challenge(Stream channel)
        {
            byte[] challenge = new byte[8];
            challenge[0] = 1;
            channel.Write(challenge, 0, challenge.Length);
            channel.Flush();

            byte[] response = YamuxSession.ReadExact(channel, challenge.Length);
            for (int i = 0; i < challenge.Length; i++)
            {
                if (challenge[i] != response[i])
                {
                    throw new InvalidDataException("challenge failed");
                }
            }
        }
    }
}
'@

Add-Type -TypeDefinition $source -Language CSharp

if ($CompileOnly) {
    [Grdp2Tcp.PowerShellServer.Program]::CompileOnly()
    return
}

[Grdp2Tcp.PowerShellServer.Program]::Run($ChannelName, $SftpServerPath, [bool]$NoSftp)
