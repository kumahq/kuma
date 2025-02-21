import{d as R,r as t,o as l,q as c,w as o,b as s,m as b,e as p,p as k,R as A,B as x,X as D,t as _,I as B,s as L}from"./index-DxWkv34s.js";import{S as N}from"./SummaryView-DD4ymD2U.js";const H=R({__name:"HostnameGeneratorListView",setup(q){return(G,r)=>{const u=t("RouteTitle"),h=t("XI18n"),g=t("XAction"),d=t("XActionGroup"),w=t("RouterView"),f=t("DataCollection"),z=t("DataLoader"),C=t("XCard"),y=t("AppView"),V=t("RouteView");return l(),c(V,{name:"hostname-generator-list-view",params:{name:"",page:1,size:15}},{default:o(({route:n,t:m,can:X,uri:v,me:i})=>[s(y,{docs:m("hostname-generators.href.docs")},{title:o(()=>[b("h1",null,[s(u,{title:m("hostname-generators.routes.items.title")},null,8,["title"])])]),default:o(()=>[r[2]||(r[2]=p()),s(h,{path:"hostname-generators.routes.items.intro","default-path":"common.i18n.ignore-error"}),r[3]||(r[3]=p()),s(C,null,{default:o(()=>[s(z,{src:v(k(A),"/hostname-generators",{},{page:n.params.page,size:n.params.size})},{loadable:o(({data:e})=>[s(f,{type:"hostname-generators",items:(e==null?void 0:e.items)??[void 0],page:n.params.page,"page-size":n.params.size,total:e==null?void 0:e.total,onChange:n.update},{default:o(()=>[s(x,{"data-testid":"hostname-generator-collection",headers:[{...i.get("headers.name"),label:m("hostname-generators.common.name"),key:"name"},{...i.get("headers.namespace"),label:m("hostname-generators.common.namespace"),key:"namespace"},...X("use zones")?[{...i.get("headers.zone"),label:m("hostname-generators.common.zone"),key:"zone"}]:[],{...i.get("headers.actions"),label:m("hostname-generators.common.actions"),key:"actions",hideLabel:!0}],items:e==null?void 0:e.items,"is-selected-row":a=>a.name===n.params.name,onResize:i.set},{name:o(({row:a})=>[s(D,{text:a.name},{default:o(()=>[s(g,{"data-action":"",to:{name:"hostname-generator-summary-view",params:{name:a.id},query:{page:n.params.page,size:n.params.size}}},{default:o(()=>[p(_(a.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),actions:o(({row:a})=>[s(d,null,{default:o(()=>[s(g,{to:{name:"hostname-generator-detail-view",params:{name:a.id}}},{default:o(()=>[p(_(m("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),r[1]||(r[1]=p()),e!=null&&e.items&&n.params.name?(l(),c(w,{key:0},{default:o(a=>[s(N,{onClose:I=>n.replace({name:"hostname-generator-list-view",params:{name:""},query:{page:n.params.page,size:n.params.size}})},{default:o(()=>[(l(),c(B(a.Component),{items:e==null?void 0:e.items},null,8,["items"]))]),_:2},1032,["onClose"])]),_:2},1024)):L("",!0)]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["docs"])]),_:1})}}});export{H as default};
