import{d as h,r as l,o as i,g as n,w as t,h as o,m as v,l as p,E as S,C as z,k as C,q as V}from"./index-622cbb72.js";import{D as k,K as q}from"./KFilterBar-5e93892b.js";import"./dataplane-0a086c06.js";const T=h({__name:"GatewayListView",setup(x){return(B,D)=>{const u=l("RouteTitle"),_=l("KSelect"),g=l("KCard"),y=l("AppView"),c=l("DataSource"),d=l("RouteView");return i(),n(c,{src:"/me"},{default:t(({data:m})=>[m?(i(),n(d,{key:0,name:"gateways-list-view",params:{page:1,size:m.pageSize,gatewayType:"all",query:"",s:"",mesh:""}},{default:t(({route:e,can:w,t:f})=>[o(c,{src:`/meshes/${e.params.mesh}/gateways/of/${e.params.gatewayType}?page=${e.params.page}&size=${e.params.size}&search=${e.params.s}`},{default:t(({data:s,error:r})=>[o(y,null,{title:t(()=>[v("h2",null,[o(u,{title:f("gateways.routes.items.title"),render:!0},null,8,["title"])])]),default:t(()=>[p(),o(g,null,{body:t(()=>[r!==void 0?(i(),n(S,{key:0,error:r},null,8,["error"])):(i(),n(k,{key:1,"data-testid":"gateway-collection",class:"gateway-collection","page-number":parseInt(e.params.page),"page-size":parseInt(e.params.size),total:s==null?void 0:s.total,items:s==null?void 0:s.items,error:r,gateways:!0,onChange:({page:a,size:b})=>{e.update({page:String(a),size:String(b)})}},{toolbar:t(()=>[o(q,{class:"data-plane-proxy-filter",placeholder:"tag: 'kuma.io/protocol: http'",query:e.params.query,fields:{name:{description:"filter by name or parts of a name"},service:{description:"filter by “kuma.io/service” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},...w("use zones")?{zone:{description:"filter by “kuma.io/zone” value"}}:{}},onFieldsChange:a=>e.update({query:a.query,s:a.query.length>0?JSON.stringify(a.fields):""})},null,8,["placeholder","query","fields","onFieldsChange"]),p(),o(_,{label:"Type","overlay-label":!0,items:[{label:"All",value:"all"},{label:"Builtin",value:"builtin"},{label:"Delegated",value:"delegated"}].map(a=>({...a,selected:a.value===e.params.gatewayType})),appearance:"select",onSelected:a=>e.update({gatewayType:String(a.value)})},{"item-template":t(({item:a})=>[p(z(a.label),1)]),_:2},1032,["items","onSelected"])]),_:2},1032,["page-number","page-size","total","items","error","onChange"]))]),_:2},1024)]),_:2},1024)]),_:2},1032,["src"])]),_:2},1032,["params"])):C("",!0)]),_:1})}}});const L=V(T,[["__scopeId","data-v-4c8142bb"]]);export{L as default};
