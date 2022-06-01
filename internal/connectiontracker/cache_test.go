package connectiontracker

import (
	"context"
	"fmt"
	"github.com/eko/gocache/v3/cache"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
	"time"
)

func Test_connCache_getOrSet(t *testing.T) {
	ctx := context.Background()
	cacheManager := newCacheManager(1 * time.Second)
	srcIP := net.ParseIP("172.217.16.14")
	dstIP := net.ParseIP("192.217.16.14")
	conn := &ConnEntry{
		SrcIP: &srcIP,
		DstIP: &dstIP,
		Ports: map[int]bool{
			9090: true,
		},
	}

	srcIPexists := net.ParseIP("172.217.16.15")
	dstIPexists := net.ParseIP("192.217.16.15")
	exists := &ConnEntry{
		SrcIP: &srcIPexists,
		DstIP: &dstIPexists,
		Ports: map[int]bool{
			9090: true,
		},
	}
	keyExists := fmt.Sprintf("%s->%s", exists.SrcIP, exists.DstIP)
	err := cacheManager.manager.Set(ctx, keyExists, exists)
	require.NoError(t, err)
	type fields struct {
		manager  *cache.Cache[*ConnEntry]
		cacheTTL time.Duration
	}
	type args struct {
		ctx    context.Context
		conn   *ConnEntry
		expire bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *ConnEntry
	}{
		{
			name: "getOrSet success",
			fields: fields{
				manager:  cacheManager.manager,
				cacheTTL: 1 * time.Second,
			},
			args: args{
				ctx:    ctx,
				conn:   conn,
				expire: false,
			},
			want: conn,
		},
		{
			name: "getOrSet already exists",
			fields: fields{
				manager:  cacheManager.manager,
				cacheTTL: 5 * time.Second,
			},
			args: args{
				ctx:    ctx,
				conn:   exists,
				expire: false,
			},
			want: exists,
		},
		{
			name: "getOrSet expire",
			fields: fields{
				manager:  cacheManager.manager,
				cacheTTL: 3 * time.Millisecond,
			},
			args: args{
				ctx:    ctx,
				conn:   conn,
				expire: true,
			},
			want: conn,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &connCache{
				manager:  tt.fields.manager,
				cacheTTL: tt.fields.cacheTTL,
			}
			if got := c.getOrSet(tt.args.ctx, tt.args.conn); !cmp.Equal(got, tt.want) {
				t.Errorf("getOrSet() = %v, want %v", got, tt.want)
				if tt.args.expire {
					key := fmt.Sprintf("%s->%s", conn.SrcIP, conn.DstIP)
					time.Sleep(5 * time.Millisecond)
					get, err := c.manager.Get(tt.args.ctx, key)
					assert.Nil(t, get)
					require.NoError(t, err)
				}
			}
		})
	}
}

func Test_updatePorts(t *testing.T) {
	srcIP := net.ParseIP("172.217.16.15")
	dstIP := net.ParseIP("192.217.16.15")
	entry1 := &ConnEntry{
		SrcIP: &srcIP,
		DstIP: &dstIP,
		Ports: map[int]bool{
			9090: true,
			8080: true,
			7070: true,
		},
	}
	entry2 := &ConnEntry{
		SrcIP: &srcIP,
		DstIP: &dstIP,
		Ports: map[int]bool{
			9090: true,
			8080: true,
			5050: true,
			6060: true,
		},
	}

	expected := &ConnEntry{
		SrcIP: &srcIP,
		DstIP: &dstIP,
		Ports: map[int]bool{
			9090: true,
			8080: true,
			7070: true,
			5050: true,
			6060: true,
		},
	}
	actual := updatePorts(entry1, entry2)
	assert.Equal(t, expected, actual)
}
