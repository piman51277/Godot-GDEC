# Godot GDEC Decoder/Encoder

A proof of concept decoder/encoder for Godot's GDEC (Encrypted) files. Currently supports v3.6 and below.

The included `dcry_v3.go` decodes and encodes GDEC files created through `File.open-encrypted-with-pass`, but can be easily modified to support `open-encrypted` as well.

Links to Godot documentation:

- [File.open-encrypted-with-pass](https://docs.godotengine.org/en/3.6/classes/class_file.html#class-file-method-open-encrypted-with-pass)
- [File.open-encrypted](https://docs.godotengine.org/en/3.6/classes/class_file.html#class-file-method-open-encrypted)

## GDEC File Format

| Field          | Type       | Value        | Description                    |
| -------------- | ---------- | ------------ | ------------------------------ |
| Magic Number   | `uint32`   | `0x47444543` | "GDEC" in ASCII                |
| File Mode      | `uint32`   | `0x00000001` | Always 1                       |
| Plaintext Hash | `byte[16]` | -            | MD5 hash of the plaintext file |
| Data Length    | `uint64`   | -            | Length of the encrypted data   |
| Data           | `byte[]`   | -            | Encrypted data                 |

## Encoding Process

1. Write magic number `0x47444543`

Found [here](https://github.com/godotengine/godot/blob/0ee0fa42e6639b6fa474b7cf6afc6b1a78142185/core/io/file_access_encrypted.cpp#L40)

2. Write file mode `0x00000001`

This is a member of Enum `FileAccessEncrypted.Mode` [link to source](https://github.com/godotengine/godot/blob/0ee0fa42e6639b6fa474b7cf6afc6b1a78142185/core/io/file_access_encrypted.h#L38). For encrypted files written to disk, only `MODE_WRITE_AES256` (`1`) is used.

3. Write MD5 hash of the plaintext file

This is the hash of the plaintext file that was encrypted. Plaintext is not padded or modified in any way before hashing.

4. Write the length of the encrypted data

This is the length of the plaintext in bytes.

5. Create an `AES-256-EBC` cipher with the provided key (method to get this key from a password is below)

6. Encrypt the plaintext data with the cipher in blocks of 16 bytes. Pad the plaintext with `0x00` to the next multiple of 16 if necessary.

## Decoding Process

This is just the reverse of the encoding process. Additional check of the MD5 hash to verify integrity is recommended.

## Deriving AES-256 key from password

See [here](https://github.com/godotengine/godot/blob/0ee0fa42e6639b6fa474b7cf6afc6b1a78142185/core/io/file_access_encrypted.cpp#L101) for Godot implementation

1. Hash the password with MD5
2. Convert to a hexademical string
3. Key is the 32 chars of the hex string
