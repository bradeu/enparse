import { expect } from "chai";
import { ethers } from "hardhat";
import type { HardhatEthersSigner } from "@nomicfoundation/hardhat-ethers/signers";
import type { IdentityRegistry } from "../typechain-types";

interface Fixture {
  registry: IdentityRegistry;
  deployer: HardhatEthersSigner;
  user1: HardhatEthersSigner;
  user2: HardhatEthersSigner;
  stranger: HardhatEthersSigner;
}

async function deployFixture(): Promise<Fixture> {
  const [deployer, user1, user2, stranger] = await ethers.getSigners();

  const IdentityRegistry = await ethers.getContractFactory("IdentityRegistry");
  const registry = await IdentityRegistry.connect(deployer).deploy() as unknown as IdentityRegistry;

  return { registry, deployer, user1, user2, stranger };
}

describe("IdentityRegistry", function () {
  it("register() stores pubkey and isRegistered returns true", async function () {
    const { registry, user1 } = await deployFixture();
    const pubkey = ethers.hexlify(ethers.randomBytes(32));
    await registry.connect(user1).register(pubkey);
    expect(await registry.isRegistered(user1.address)).to.be.true;
    expect(await registry.getPublicKey(user1.address)).to.equal(pubkey);
  });

  it("isRegistered returns false for unregistered address", async function () {
    const { registry, stranger } = await deployFixture();
    expect(await registry.isRegistered(stranger.address)).to.be.false;
  });

  it("getPublicKey returns zero bytes32 for unregistered address", async function () {
    const { registry, stranger } = await deployFixture();
    expect(await registry.getPublicKey(stranger.address)).to.equal(ethers.ZeroHash);
  });

  it("register() rejects zero bytes32 pubkey", async function () {
    const { registry, user1 } = await deployFixture();
    await expect(
      registry.connect(user1).register(ethers.ZeroHash)
    ).to.be.revertedWith("Invalid public key");
  });

  it("register() emits Registered event", async function () {
    const { registry, user1 } = await deployFixture();
    const pubkey = ethers.hexlify(ethers.randomBytes(32));
    await expect(registry.connect(user1).register(pubkey))
      .to.emit(registry, "Registered")
      .withArgs(user1.address, pubkey);
  });

  it("allows a user to update their pubkey with a second register()", async function () {
    const { registry, user1 } = await deployFixture();
    const key1 = ethers.hexlify(ethers.randomBytes(32));
    const key2 = ethers.hexlify(ethers.randomBytes(32));
    await registry.connect(user1).register(key1);
    await registry.connect(user1).register(key2);
    expect(await registry.getPublicKey(user1.address)).to.equal(key2);
  });

  it("each address stores an independent pubkey", async function () {
    const { registry, user1, user2 } = await deployFixture();
    const key1 = ethers.hexlify(ethers.randomBytes(32));
    const key2 = ethers.hexlify(ethers.randomBytes(32));
    await registry.connect(user1).register(key1);
    await registry.connect(user2).register(key2);
    expect(await registry.getPublicKey(user1.address)).to.equal(key1);
    expect(await registry.getPublicKey(user2.address)).to.equal(key2);
  });

  it("registering user1 does not affect isRegistered for user2", async function () {
    const { registry, user1, user2 } = await deployFixture();
    await registry.connect(user1).register(ethers.hexlify(ethers.randomBytes(32)));
    expect(await registry.isRegistered(user2.address)).to.be.false;
  });
});
