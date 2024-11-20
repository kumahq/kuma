import{d as b,e as n,o as i,m as d,w as a,a as o,b as r,l as R,aq as D,A as S,a0 as f,t as c,c as _,J as v,S as P,F as T,p as B}from"./index-B-FVJ5xI.js";import{S as L}from"./SummaryView-DtrX1zdJ.js";const G=b({__name:"ServiceListView",setup(N){return(X,p)=>{const h=n("RouteTitle"),u=n("XAction"),w=n("XActionGroup"),y=n("RouterView"),C=n("DataCollection"),k=n("DataLoader"),z=n("KCard"),V=n("AppView"),A=n("RouteView");return i(),d(A,{name:"service-list-view",params:{page:1,size:50,mesh:"",service:""}},{default:a(({route:s,t:m,uri:x,me:l})=>[o(h,{render:!1,title:m("services.routes.items.title")},null,8,["title"]),p[5]||(p[5]=r()),o(V,{docs:m("services.href.docs")},{default:a(()=>[o(z,null,{default:a(()=>[o(k,{src:x(R(D),"/meshes/:mesh/service-insights/of/:serviceType",{mesh:s.params.mesh,serviceType:"internal"},{page:s.params.page,size:s.params.size})},{loadable:a(({data:t})=>[o(C,{type:"services",items:(t==null?void 0:t.items)??[void 0],page:s.params.page,"page-size":s.params.size,total:t==null?void 0:t.total,onChange:s.update},{default:a(()=>[o(S,{class:"service-collection","data-testid":"service-collection",headers:[{...l.get("headers.name"),label:"Name",key:"name"},{...l.get("headers.addressPort"),label:"Address",key:"addressPort"},{...l.get("headers.online"),label:"DP proxies (online / total)",key:"online"},{...l.get("headers.status"),label:"Status",key:"status"},{...l.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:t==null?void 0:t.items,"is-selected-row":e=>e.name===s.params.service,onResize:l.set},{name:a(({row:e})=>[o(f,{text:e.name},{default:a(()=>[o(u,{"data-action":"",to:{name:"service-detail-view",params:{mesh:e.mesh,service:e.name},query:{page:s.params.page,size:s.params.size}}},{default:a(()=>[r(c(e.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),addressPort:a(({row:e})=>[e.addressPort?(i(),d(f,{key:0,text:e.addressPort},null,8,["text"])):(i(),_(v,{key:1},[r(c(m("common.collection.none")),1)],64))]),online:a(({row:e})=>[e.dataplanes?(i(),_(v,{key:0},[r(c(e.dataplanes.online||0)+" / "+c(e.dataplanes.total||0),1)],64)):(i(),_(v,{key:1},[r(c(m("common.collection.none")),1)],64))]),status:a(({row:e})=>[o(P,{status:e.status},null,8,["status"])]),actions:a(({row:e})=>[o(w,null,{default:a(()=>[o(u,{to:{name:"service-detail-view",params:{mesh:e.mesh,service:e.name}}},{default:a(()=>[r(c(m("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),p[4]||(p[4]=r()),s.params.service?(i(),d(y,{key:0},{default:a(e=>[o(L,{onClose:g=>s.replace({name:"service-list-view",params:{mesh:s.params.mesh},query:{page:s.params.page,size:s.params.size}})},{default:a(()=>[(i(),d(T(e.Component),{name:s.params.service,service:t==null?void 0:t.items.find(g=>g.name===s.params.service)},null,8,["name","service"]))]),_:2},1032,["onClose"])]),_:2},1024)):B("",!0)]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["docs"])]),_:1})}}});export{G as default};
