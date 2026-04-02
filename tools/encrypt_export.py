import argparse
import os

from cryptography.hazmat.primitives.ciphers.aead import AESGCM
from cryptography.hazmat.primitives.kdf.scrypt import Scrypt


MAGIC = b"PHX1"  


def derive_key(passphrase: str, salt: bytes) -> bytes:
    kdf = Scrypt(
        salt=salt,
        length=32,
        n=2**15,
        r=8,
        p=1,
    )
    return kdf.derive(passphrase.encode("utf-8"))


def encrypt_bytes(plaintext: bytes, passphrase: str) -> bytes:
    salt = os.urandom(16)
    nonce = os.urandom(12)
    key = derive_key(passphrase, salt)
    aead = AESGCM(key)
    ct = aead.encrypt(nonce, plaintext, None)
    # Format:
    # MAGIC(4) | salt(16) | nonce(12) | ciphertext+tag(variable)
    return MAGIC + salt + nonce + ct


def decrypt_bytes(blob: bytes, passphrase: str) -> bytes:
    if len(blob) < 4 + 16 + 12:
        raise ValueError("ciphertext terlalu pendek")
    if blob[:4] != MAGIC:
        raise ValueError("format file tidak valid")
    salt = blob[4:20]
    nonce = blob[20:32]
    ct = blob[32:]
    key = derive_key(passphrase, salt)
    aead = AESGCM(key)
    return aead.decrypt(nonce, ct, None)


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--mode", choices=["encrypt", "decrypt"], default="encrypt")
    ap.add_argument("--in", dest="in_path", required=True)
    ap.add_argument("--out", dest="out_path", required=True)
    ap.add_argument("--passphrase", required=True)
    args = ap.parse_args()

    with open(args.in_path, "rb") as f:
        src = f.read()

    if args.mode == "encrypt":
        out_bytes = encrypt_bytes(src, args.passphrase)
    else:
        out_bytes = decrypt_bytes(src, args.passphrase)

    with open(args.out_path, "wb") as f:
        f.write(out_bytes)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

