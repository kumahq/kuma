import{d as A,r as n,o as i,p as _,w as a,b as s,m as C,as as x,A as b,V as u,e as p,t as c,c as r,J as m,S as V}from"./index-yoi81zLz.js";const R=A({__name:"DelegatedGatewayListView",setup(D){return(P,B)=>{const g=n("XAction"),h=n("XActionGroup"),y=n("DataCollection"),w=n("DataLoader"),f=n("KCard"),k=n("AppView"),v=n("RouteView");return i(),_(v,{name:"delegated-gateway-list-view",params:{page:1,size:50,mesh:""}},{default:a(({route:o,t:d,me:l,uri:z})=>[s(k,{docs:d("delegated-gateways.href.docs")},{default:a(()=>[s(f,null,{default:a(()=>[s(w,{src:z(C(x),"/meshes/:mesh/service-insights/of/:serviceType",{mesh:o.params.mesh,serviceType:"gateway_delegated"},{page:o.params.page,size:o.params.size})},{loadable:a(({data:t})=>[s(y,{type:"gateways",items:(t==null?void 0:t.items)??[void 0],page:o.params.page,"page-size":o.params.size,total:t==null?void 0:t.total,onChange:o.update},{default:a(()=>[s(b,{class:"delegated-gateway-collection","data-testid":"delegated-gateway-collection",headers:[{...l.get("headers.name"),label:"Name",key:"name"},{...l.get("headers.addressPort"),label:"Address",key:"addressPort"},{...l.get("headers.dataplanes"),label:"DP proxies (online / total)",key:"dataplanes"},{...l.get("headers.status"),label:"Status",key:"status"},{...l.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:t==null?void 0:t.items,onResize:l.set},{name:a(({row:e})=>[s(u,{text:e.name},{default:a(()=>[s(g,{to:{name:"delegated-gateway-detail-view",params:{mesh:e.mesh,service:e.name},query:{page:o.params.page,size:o.params.size}}},{default:a(()=>[p(c(e.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),addressPort:a(({row:e})=>[e.addressPort?(i(),_(u,{key:0,text:e.addressPort},null,8,["text"])):(i(),r(m,{key:1},[p(c(d("common.collection.none")),1)],64))]),dataplanes:a(({row:e})=>[e.dataplanes?(i(),r(m,{key:0},[p(c(e.dataplanes.online||0)+" / "+c(e.dataplanes.total||0),1)],64)):(i(),r(m,{key:1},[p(c(d("common.collection.none")),1)],64))]),status:a(({row:e})=>[s(V,{status:e.status},null,8,["status"])]),actions:a(({row:e})=>[s(h,null,{default:a(()=>[s(g,{to:{name:"delegated-gateway-detail-view",params:{mesh:e.mesh,service:e.name}}},{default:a(()=>[p(c(d("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","onResize"])]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["docs"])]),_:1})}}});export{R as default};
