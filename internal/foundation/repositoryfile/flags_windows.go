package repositoryfile

// Rooted traversal and descriptor identity checks protect Windows paths;
// Unix FIFO and O_NOFOLLOW flags do not exist on this platform.
const openSafetyFlags = 0
