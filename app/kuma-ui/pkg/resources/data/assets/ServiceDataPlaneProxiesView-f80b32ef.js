import{D as z,K as q}from"./KFilterBar-79182a3a.js";import{S as T}from"./SummaryView-227f39d3.js";import{d as $,r as s,o,i as l,w as t,j as i,p as B,n,l as f,F as w,I as K,H as R,m as p,q as F,t as N}from"./index-c6bd05ee.js";import"./AppCollection-6aabd095.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-fb6e10fa.js";import"./StatusBadge-07ca9e6a.js";import"./dataplane-0a086c06.js";const I=$({__name:"ServiceDataPlaneProxiesView",setup(A){return(L,j)=>{const h=s("RouteTitle"),b=s("KSelect"),S=s("KCard"),C=s("RouterView"),c=s("DataSource"),V=s("AppView"),x=s("RouteView");return o(),l(c,{src:"/me"},{default:t(({data:d})=>[d?(o(),l(x,{key:0,name:"service-data-plane-proxies-view",params:{page:1,size:d.pageSize,query:"",s:"",mesh:"",service:"",gatewayType:"",dataPlane:""}},{default:t(({route:e,t:P})=>[i(V,null,{title:t(()=>[B("h2",null,[i(h,{title:P("services.routes.item.navigation.service-data-plane-proxies-view"),render:!0},null,8,["title"])])]),default:t(()=>[n(),i(c,{src:`/meshes/${e.params.mesh}/dataplanes/for/${e.params.service}/of/all?page=${e.params.page}&size=${e.params.size}&search=${e.params.s}`},{default:t(({data:r,error:k})=>{var u,_,y,g;return[(o(!0),f(w,null,K([((g=(y=(_=(u=r==null?void 0:r.items)==null?void 0:u[0])==null?void 0:_.dataplane)==null?void 0:y.networking)==null?void 0:g.gateway)!==void 0],m=>(o(),f(w,{key:m},[i(S,null,{body:t(()=>[i(z,{"data-testid":"data-plane-collection",class:"data-plane-collection","page-number":parseInt(e.params.page),"page-size":parseInt(e.params.size),total:r==null?void 0:r.total,items:r==null?void 0:r.items,error:k,gateways:m,"is-selected-row":a=>a.name===e.params.dataPlane,"summary-route-name":"service-data-plane-summary-view",onChange:e.update},{toolbar:t(()=>[i(q,{class:"data-plane-proxy-filter",placeholder:"tag: 'kuma.io/protocol: http'",query:e.params.query,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},zone:{description:"filter by “kuma.io/zone” value"}},onFieldsChange:a=>e.update({query:a.query,s:a.query.length>0?JSON.stringify(a.fields):""})},null,8,["placeholder","query","fields","onFieldsChange"]),n(),m?(o(),l(b,{key:0,label:"Type","overlay-label":!0,items:[{label:"All",value:"all"},{label:"Builtin",value:"builtin"},{label:"Delegated",value:"delegated"}].map(a=>({...a,selected:a.value===e.params.gatewayType})),appearance:"select",onSelected:a=>e.update({gatewayType:String(a.value)})},{"item-template":t(({item:a})=>[n(R(a.label),1)]),_:2},1032,["items","onSelected"])):p("",!0)]),_:2},1032,["page-number","page-size","total","items","error","gateways","is-selected-row","onChange"])]),_:2},1024),n(),e.params.dataPlane?(o(),l(C,{key:0},{default:t(a=>[i(T,{onClose:v=>e.replace({name:"service-data-plane-proxies-view",params:{mesh:e.params.mesh},query:{page:e.params.page,size:e.params.size}})},{default:t(()=>[(o(),l(F(a.Component),{name:e.params.dataPlane,"dataplane-overview":r==null?void 0:r.items.find(v=>v.name===e.params.dataPlane)},null,8,["name","dataplane-overview"]))]),_:2},1032,["onClose"])]),_:2},1024)):p("",!0)],64))),128))]}),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["params"])):p("",!0)]),_:1})}}});const U=N(I,[["__scopeId","data-v-5c3346f3"]]);export{U as default};
