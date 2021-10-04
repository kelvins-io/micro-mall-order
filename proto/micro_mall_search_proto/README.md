# micro-mall-search-proto

#### 介绍
搜索服务-proto仓库

#### 软件架构
软件架构说明


#### 使用说明
接口定义
```protobuf
    // 商品库存搜索
    rpc SkuInventorySearch(SkuInventorySearchRequest)returns(SkuInventorySearchResponse) {
        option (google.api.http) = {
            get: "/v1/search/sku_inventory"
        };
    }
    // 店铺搜索
    rpc ShopSearch(ShopSearchRequest) returns (ShopSearchResponse) {
        option (google.api.http) = {
            get: "/v1/search/shop"
        };
    }
    // 商品库存同步
    rpc SyncSkuInventory(SyncSkuInventoryRequest)returns(SyncSkuInventoryResponse) {
        option (google.api.http) = {
            post: "/v1/search/sync/sku_inventory"
            body:"*"
        };
    }
    // 店铺同步
    rpc SyncShop(SyncShopRequest)returns(SyncShopResponse) {
        option (google.api.http) = {
            post: "/v1/search/sync/shop"
            body:"*"
        };
    }
```

#### 参与贡献

1.  Fork 本仓库
2.  新建 Feat_xxx 分支
3.  提交代码
4.  新建 Pull Request
