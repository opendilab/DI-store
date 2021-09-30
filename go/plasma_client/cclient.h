#ifndef _CCLIENT_H_
#define _CCLIENT_H_
#include <stdint.h>

#ifdef __cplusplus
extern "C"
{
#endif
    typedef void *ClientPointer;
    typedef void *ObjectBufferPointer;

    typedef struct
    {
        const unsigned char *data;
        int64_t size;
    } Data;

    typedef struct
    {
        const unsigned char *data;
        int64_t size;
        void *inner;
    } Buff;

    ClientPointer NewClient();

    void DeleteClient(ClientPointer client);

    const char *Connect(ClientPointer client, const char *path);

    const char *Disconnect(ClientPointer client);

    const char *Create(ClientPointer client, const char *object_id, int len, Buff **ppbuff);

    void Abort(ClientPointer client, const char *object_id);

    const char *Seal(ClientPointer client, const char *object_id);

    const char *Get(ClientPointer client, char *object_id, int64_t timeout_ms, Buff **ppbuff);

    void DeleteBuff(Buff *pbuff);

    // const char *CreateAndSeal(ClientPointer client, const char *object_id, const char *data,
    //                           int len, bool evict_if_full, int memcopy_threads,
    //                           int64_t memcopy_threshold, int64_t memcopy_blocksize);

    // const char *Get(ClientPointer client, char **object_ids, int64_t num_objects,
    //                 int64_t timeout_ms, ObjectBufferPointer *object_buffer);

    // Data GetData(ObjectBufferPointer object_buffers, unsigned int index);

    // unsigned int DataSize(ObjectBufferPointer object_buffers);

    // void DeleteObjectBufferPointer(ObjectBufferPointer object_buffers);

    const char *Delete(ClientPointer client, char **object_ids, int64_t num_objects);

    const char *Contains(ClientPointer client, char *object_id, char *has_object);

    void ParallelMemCopy(void *dst, void *src, int64_t nbytes, uintptr_t block_size, int num_threads);

    void ParallelMemCopy2(void *dst, void *src, int64_t nbytes, uintptr_t block_size, int num_threads);

#ifdef __cplusplus
}
#endif // extern "C"

#endif // #ifdef _CCLIENT_H_
