import{d as A,r as o,o as i,m as p,w as a,b as n,e as r,l as x,aF as R,A as D,T as h,t as l,c as d,F as _,S,E as P,p as T}from"./index-DxrN05KS.js";import{S as B}from"./SummaryView-jGgfzaaj.js";const E=A({__name:"ServiceListView",setup(L){return(N,X)=>{const g=o("RouteTitle"),u=o("XAction"),f=o("XActionGroup"),w=o("RouterView"),y=o("DataCollection"),C=o("DataLoader"),k=o("KCard"),z=o("AppView"),V=o("RouteView");return i(),p(V,{name:"service-list-view",params:{page:1,size:50,mesh:"",service:""}},{default:a(({route:s,t:m,uri:b,me:c})=>[n(g,{render:!1,title:m("services.routes.items.title")},null,8,["title"]),r(),n(z,{docs:m("services.href.docs")},{default:a(()=>[n(k,null,{default:a(()=>[n(C,{src:b(x(R),"/meshes/:mesh/service-insights/of/:serviceType",{mesh:s.params.mesh,serviceType:"internal"},{page:s.params.page,size:s.params.size})},{loadable:a(({data:t})=>[n(y,{type:"services",items:(t==null?void 0:t.items)??[void 0]},{default:a(()=>[n(D,{class:"service-collection","data-testid":"service-collection",headers:[{...c.get("headers.name"),label:"Name",key:"name"},{...c.get("headers.addressPort"),label:"Address",key:"addressPort"},{...c.get("headers.online"),label:"DP proxies (online / total)",key:"online"},{...c.get("headers.status"),label:"Status",key:"status"},{...c.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],"page-number":s.params.page,"page-size":s.params.size,total:t==null?void 0:t.total,items:t==null?void 0:t.items,"is-selected-row":e=>e.name===s.params.service,onChange:s.update,onResize:c.set},{name:a(({row:e})=>[n(h,{text:e.name},{default:a(()=>[n(u,{"data-action":"",to:{name:"service-detail-view",params:{mesh:e.mesh,service:e.name},query:{page:s.params.page,size:s.params.size}}},{default:a(()=>[r(l(e.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),addressPort:a(({row:e})=>[e.addressPort?(i(),p(h,{key:0,text:e.addressPort},null,8,["text"])):(i(),d(_,{key:1},[r(l(m("common.collection.none")),1)],64))]),online:a(({row:e})=>[e.dataplanes?(i(),d(_,{key:0},[r(l(e.dataplanes.online||0)+" / "+l(e.dataplanes.total||0),1)],64)):(i(),d(_,{key:1},[r(l(m("common.collection.none")),1)],64))]),status:a(({row:e})=>[n(S,{status:e.status},null,8,["status"])]),actions:a(({row:e})=>[n(f,null,{default:a(()=>[n(u,{to:{name:"service-detail-view",params:{mesh:e.mesh,service:e.name}}},{default:a(()=>[r(l(m("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","page-number","page-size","total","items","is-selected-row","onChange","onResize"]),r(),s.params.service?(i(),p(w,{key:0},{default:a(e=>[n(B,{onClose:v=>s.replace({name:"service-list-view",params:{mesh:s.params.mesh},query:{page:s.params.page,size:s.params.size}})},{default:a(()=>[(i(),p(P(e.Component),{name:s.params.service,service:t==null?void 0:t.items.find(v=>v.name===s.params.service)},null,8,["name","service"]))]),_:2},1032,["onClose"])]),_:2},1024)):T("",!0)]),_:2},1032,["items"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["docs"])]),_:1})}}});export{E as default};
