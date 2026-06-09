import { expect } from "chai";
import { ethers } from "hardhat";
import type { HardhatEthersSigner } from "@nomicfoundation/hardhat-ethers/signers";
import type { ProjectVault, IdentityRegistry } from "../typechain-types";

interface Fixture {
  registry: IdentityRegistry;
  vault: ProjectVault;
  deployer: HardhatEthersSigner;
  owner: HardhatEthersSigner;
  member1: HardhatEthersSigner;
  member2: HardhatEthersSigner;
  stranger: HardhatEthersSigner;
}

async function deployFixture(): Promise<Fixture> {
  const [deployer, owner, member1, member2, stranger] = await ethers.getSigners();

  const IdentityRegistry = await ethers.getContractFactory("IdentityRegistry");
  const registry = await IdentityRegistry.connect(deployer).deploy() as unknown as IdentityRegistry;
  const registryAddr = await registry.getAddress();

  const ProjectVault = await ethers.getContractFactory("ProjectVault");
  const vault = await ProjectVault.connect(deployer).deploy(registryAddr) as unknown as ProjectVault;

  await registry.connect(owner).register(ethers.hexlify(ethers.randomBytes(32)));
  await registry.connect(member1).register(ethers.hexlify(ethers.randomBytes(32)));
  await registry.connect(member2).register(ethers.hexlify(ethers.randomBytes(32)));

  return { registry, vault, deployer, owner, member1, member2, stranger };
}

function projectId(name: string): string {
  return ethers.keccak256(ethers.toUtf8Bytes(name));
}

// ─── createProject ───────────────────────────────────────────────────────────

describe("ProjectVault — createProject", function () {
  it("sets owner and adds owner as first member", async function () {
    const { vault, owner } = await deployFixture();
    await vault.connect(owner).createProject("alpha");
    const id = projectId("alpha");
    expect(await vault.getOwner(id)).to.equal(owner.address);
    expect(await vault.isMember(id, owner.address)).to.be.true;
  });

  it("stores project name in projectNames mapping", async function () {
    const { vault, owner } = await deployFixture();
    await vault.connect(owner).createProject("alpha");
    expect(await vault.projectNames(projectId("alpha"))).to.equal("alpha");
  });

  it("appends projectId to owner's memberProjects list", async function () {
    const { vault, owner } = await deployFixture();
    await vault.connect(owner).createProject("alpha");
    const ids = await vault.getMemberProjects(owner.address);
    expect(ids).to.include(projectId("alpha"));
  });

  it("emits ProjectCreated with correct args", async function () {
    const { vault, owner } = await deployFixture();
    const id = projectId("alpha");
    await expect(vault.connect(owner).createProject("alpha"))
      .to.emit(vault, "ProjectCreated")
      .withArgs(id, owner.address, "alpha");
  });

  it("rejects empty project name", async function () {
    const { vault, owner } = await deployFixture();
    await expect(vault.connect(owner).createProject("")).to.be.revertedWith("Name required");
  });

  it("rejects duplicate project name", async function () {
    const { vault, owner } = await deployFixture();
    await vault.connect(owner).createProject("alpha");
    await expect(vault.connect(owner).createProject("alpha")).to.be.revertedWith("Project already exists");
  });

  it("rejects unregistered creator", async function () {
    const { vault, stranger } = await deployFixture();
    await expect(vault.connect(stranger).createProject("alpha")).to.be.revertedWith("Must register identity first");
  });

  it("supports multiple distinct projects", async function () {
    const { vault, owner } = await deployFixture();
    await vault.connect(owner).createProject("alpha");
    await vault.connect(owner).createProject("beta");
    expect(await vault.getOwner(projectId("alpha"))).to.equal(owner.address);
    expect(await vault.getOwner(projectId("beta"))).to.equal(owner.address);
  });
});

// ─── addMember ───────────────────────────────────────────────────────────────

describe("ProjectVault — addMember", function () {
  async function setup() {
    const fixture = await deployFixture();
    await fixture.vault.connect(fixture.owner).createProject("alpha");
    return { ...fixture, id: projectId("alpha") };
  }

  it("owner can add a registered member", async function () {
    const { vault, owner, member1, id } = await setup();
    await vault.connect(owner).addMember(id, member1.address);
    expect(await vault.isMember(id, member1.address)).to.be.true;
  });

  it("appends projectId to new member's project list", async function () {
    const { vault, owner, member1, id } = await setup();
    await vault.connect(owner).addMember(id, member1.address);
    const ids = await vault.getMemberProjects(member1.address);
    expect(ids).to.include(id);
  });

  it("emits MemberAdded", async function () {
    const { vault, owner, member1, id } = await setup();
    await expect(vault.connect(owner).addMember(id, member1.address))
      .to.emit(vault, "MemberAdded")
      .withArgs(id, member1.address);
  });

  it("non-owner cannot add a member", async function () {
    const { vault, member1, id } = await setup();
    await expect(
      vault.connect(member1).addMember(id, member1.address)
    ).to.be.revertedWith("Not project owner");
  });

  it("rejects adding an unregistered address", async function () {
    const { vault, owner, stranger, id } = await setup();
    await expect(
      vault.connect(owner).addMember(id, stranger.address)
    ).to.be.revertedWith("Member must register identity first");
  });

  it("rejects duplicate membership", async function () {
    const { vault, owner, member1, id } = await setup();
    await vault.connect(owner).addMember(id, member1.address);
    await expect(
      vault.connect(owner).addMember(id, member1.address)
    ).to.be.revertedWith("Already a member");
  });

  it("owner is already a member — re-add reverts", async function () {
    const { vault, owner, id } = await setup();
    await expect(
      vault.connect(owner).addMember(id, owner.address)
    ).to.be.revertedWith("Already a member");
  });

  it("reverts for non-existent project", async function () {
    const { vault, owner, member1 } = await setup();
    await expect(
      vault.connect(owner).addMember(projectId("does-not-exist"), member1.address)
    ).to.be.revertedWith("Project does not exist");
  });
});

// ─── setSecret / getSecret ───────────────────────────────────────────────────

describe("ProjectVault — setSecret / getSecret", function () {
  async function setup() {
    const fixture = await deployFixture();
    await fixture.vault.connect(fixture.owner).createProject("alpha");
    const id = projectId("alpha");
    await fixture.vault.connect(fixture.owner).addMember(id, fixture.member1.address);
    return { ...fixture, id };
  }

  it("each member receives their own encrypted blob", async function () {
    const { vault, owner, member1, id } = await setup();
    const ownerBlob = ethers.toUtf8Bytes("enc-for-owner");
    const member1Blob = ethers.toUtf8Bytes("enc-for-member1");
    await vault.connect(owner).setSecret(id, "DB_URL", [owner.address, member1.address], [ownerBlob, member1Blob]);
    expect(ethers.toUtf8String(await vault.connect(owner).getSecret(id, "DB_URL"))).to.equal("enc-for-owner");
    expect(ethers.toUtf8String(await vault.connect(member1).getSecret(id, "DB_URL"))).to.equal("enc-for-member1");
  });

  it("owner can read their own blob when they are the only member set", async function () {
    const { vault, owner, id } = await setup();
    const blob = ethers.toUtf8Bytes("owner-secret");
    await vault.connect(owner).setSecret(id, "KEY", [owner.address], [blob]);
    expect(ethers.toUtf8String(await vault.connect(owner).getSecret(id, "KEY"))).to.equal("owner-secret");
  });

  it("overwriting a secret updates the stored blob", async function () {
    const { vault, owner, id } = await setup();
    await vault.connect(owner).setSecret(id, "KEY", [owner.address], [ethers.toUtf8Bytes("v1")]);
    await vault.connect(owner).setSecret(id, "KEY", [owner.address], [ethers.toUtf8Bytes("v2")]);
    expect(ethers.toUtf8String(await vault.connect(owner).getSecret(id, "KEY"))).to.equal("v2");
  });

  it("non-member cannot call getSecret", async function () {
    const { vault, owner, stranger, id } = await setup();
    await vault.connect(owner).setSecret(id, "KEY", [owner.address], [ethers.toUtf8Bytes("x")]);
    await expect(vault.connect(stranger).getSecret(id, "KEY")).to.be.revertedWith("Not a project member");
  });

  it("non-owner cannot call setSecret", async function () {
    const { vault, member1, id } = await setup();
    await expect(
      vault.connect(member1).setSecret(id, "KEY", [member1.address], [ethers.toUtf8Bytes("x")])
    ).to.be.revertedWith("Not project owner");
  });

  it("rejects mismatched members/values array lengths", async function () {
    const { vault, owner, member1, id } = await setup();
    await expect(
      vault.connect(owner).setSecret(id, "KEY", [owner.address, member1.address], [ethers.toUtf8Bytes("x")])
    ).to.be.revertedWith("Length mismatch");
  });

  it("rejects empty members array", async function () {
    const { vault, owner, id } = await setup();
    await expect(vault.connect(owner).setSecret(id, "KEY", [], [])).to.be.revertedWith("No members provided");
  });

  it("tracks secret names and does not duplicate on overwrite", async function () {
    const { vault, owner, id } = await setup();
    const blob = ethers.toUtf8Bytes("x");
    await vault.connect(owner).setSecret(id, "FOO", [owner.address], [blob]);
    await vault.connect(owner).setSecret(id, "BAR", [owner.address], [blob]);
    await vault.connect(owner).setSecret(id, "FOO", [owner.address], [blob]);
    const names = await vault.connect(owner).getSecretNames(id);
    expect(names).to.include("FOO");
    expect(names).to.include("BAR");
    expect(names.filter((n: string) => n === "FOO").length).to.equal(1);
  });

  it("non-member cannot call getSecretNames", async function () {
    const { vault, owner, stranger, id } = await setup();
    await vault.connect(owner).setSecret(id, "FOO", [owner.address], [ethers.toUtf8Bytes("x")]);
    await expect(vault.connect(stranger).getSecretNames(id)).to.be.revertedWith("Not a project member");
  });

  it("re-encrypt regression: new member gets blob after owner re-calls setSecret", async function () {
    const { vault, owner, member2, id } = await setup();
    await vault.connect(owner).setSecret(id, "DB_URL", [owner.address], [ethers.toUtf8Bytes("enc-for-owner")]);
    await vault.connect(owner).addMember(id, member2.address);
    await vault.connect(owner).setSecret(
      id, "DB_URL",
      [owner.address, member2.address],
      [ethers.toUtf8Bytes("enc-for-owner-v2"), ethers.toUtf8Bytes("enc-for-member2")]
    );
    expect(ethers.toUtf8String(await vault.connect(member2).getSecret(id, "DB_URL"))).to.equal("enc-for-member2");
    expect(ethers.toUtf8String(await vault.connect(owner).getSecret(id, "DB_URL"))).to.equal("enc-for-owner-v2");
  });
});

// ─── getMembers ──────────────────────────────────────────────────────────────

describe("ProjectVault — getMembers", function () {
  it("returns owner + all added members", async function () {
    const { vault, owner, member1, member2 } = await deployFixture();
    await vault.connect(owner).createProject("alpha");
    const id = projectId("alpha");
    await vault.connect(owner).addMember(id, member1.address);
    await vault.connect(owner).addMember(id, member2.address);
    const members = await vault.getMembers(id);
    expect(members).to.include(owner.address);
    expect(members).to.include(member1.address);
    expect(members).to.include(member2.address);
  });

  it("reverts for non-existent project", async function () {
    const { vault } = await deployFixture();
    await expect(vault.getMembers(projectId("ghost"))).to.be.revertedWith("Project does not exist");
  });
});
