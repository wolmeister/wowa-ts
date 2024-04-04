const fs = require('fs');

const data = fs.readFileSync('./BetterFishing-1.1.22.zip');

// END OF CENTRAL DIRECTORY
// Bytes | Description
// ------+-------------------------------------------------------------------
//     4 | Signature (0x06054b50)
//     2 | Number of this disk
//     2 | Disk where central directory starts
//     2 | Numbers of central directory records on this disk
//     2 | Total number of central directory records
//     4 | Size of central directory in bytes
//     4 | Offset to start of central directory
//     2 | Comment length (n)
//     n | Comment
const EOCD_SIGNATURE = 0x06054b50;

function findSignatureOffset(buffer, signature) {
  for (let i = buffer.length - 4; i >= 0; i--) {
    if (data.readUInt32LE(i) === signature) {
      return i;
    }
  }
  return -1;
}

const eocdOffset = findSignatureOffset(data, EOCD_SIGNATURE);
console.log('eocd offset', eocdOffset);

const diskNumber = data.readInt16LE(eocdOffset + 4);
console.log('diskNumber', diskNumber);

// END OF CENTRAL DIRECTORY
// Bytes | Description
// ------+-------------------------------------------------------------------
//     4 | Signature (0x06054b50)
//     2 | Number of this disk
//     2 | Disk where central directory starts
//     2 | Numbers of central directory records on this disk
//     2 | Total number of central directory records
//     4 | Size of central directory in bytes
//     4 | Offset to start of central directory
//     2 | Comment length (n)
//     n | Comment
type EndOfCentralDirectory = {};

class ZipFile {
  static EOCD_SIGNATURE = 0x06054b50;

  /**
   * @param {Buffer} buffer
   */
  static parse(buffer) {
    let eocdOffset = -1;
    for (let i = buffer.length - 4; i >= 0; i--) {
      if (buffer.readUInt32LE(i) === EOCD_SIGNATURE) {
        eocdOffset = i;
        break;
      }
    }

    if (eocdOffset === -1) {
      throw new Error('Data is not a valid zip archive');
    }

    const diskNumberOffset = eocdOffset + 4;
    const cdStartDiskOffset = diskNumberOffset + 2;
    const cdDiskCountOffset = cdStartDiskOffset + 2;
    const cdTotalCountOffset = cdDiskCountOffset + 2;
    const cdSizeOffset = cdTotalCountOffset + 2;
    const cdStartOffsetOffset = cdSizeOffset + 4;
    const commentLengthOffset = cdStartOffsetOffset + 4;
    const commentOffset = commentLengthOffset + 2;

    const diskNumber = buffer.readInt16LE(diskNumberOffset);
    console.log('diskNumber', diskNumber);
    const cdStartDisk = buffer.readInt16LE(cdStartDiskOffset);
    console.log('cdStartDisk', cdStartDisk);
    const cdDiskCount = buffer.readInt16LE(cdDiskCountOffset);
    console.log('cdDiskCount', cdDiskCount);
    const cdTotalCount = buffer.readInt16LE(cdTotalCountOffset);
    console.log('cdTotalCount', cdTotalCount);
    const cdSize = buffer.readInt16LE(cdSizeOffset);
    console.log('cdSize', cdSize);
    const cdStartOffset = buffer.readInt16LE(cdStartOffsetOffset);
    console.log('cdStartOffset', cdStartOffset);
    const commentLength = buffer.readInt16LE(commentLengthOffset);
    console.log('commentLength', commentLength);
    const comment = buffer.slice(commentOffset, commentOffset + commentLength);
    console.log('comment', comment.toString());
  }
}

const zip = ZipFile.parse(data);
