import{d as N,l as h,R as D,a as c,o as t,b as i,w as o,F as f,t as b,f as l,c as g,m as x,q as B,e as d,C as v}from"./index-G4vI7xl4.js";import{N as I}from"./NavTabs-t0SuNKfL.js";const S=N({__name:"DataPlaneInboundSummaryView",props:{dataplaneType:{},gateway:{},inbounds:{}},setup(V){var w;const{t:C}=h(),R=D(),n=V,k=(((w=R.getRoutes().find(e=>e.name==="data-plane-inbound-summary-view"))==null?void 0:w.children)??[]).map(e=>{var r,a;const m=typeof e.name>"u"?(r=e.children)==null?void 0:r[0]:e,s=m.name,u=((a=m.meta)==null?void 0:a.module)??"";return{title:C(`data-planes.routes.item.navigation.${s}`),routeName:s,module:u}});return(e,m)=>{const s=c("DataCollection"),u=c("RouterView"),y=c("AppView"),r=c("RouteView");return t(),i(r,{name:"data-plane-inbound-summary-view",params:{service:""}},{default:o(({route:a})=>[d(y,null,{title:o(()=>[x("h2",null,[n.gateway?(t(),g(f,{key:0},[l(b(a.params.service),1)],64)):(t(),g(f,{key:1},[l(`
            Inbound `+b(a.params.service.replace("localhost_","")),1)],64))])]),default:o(()=>[l(),d(I,{tabs:B(k)},null,8,["tabs"]),l(),d(u,null,{default:o(_=>[n.gateway?(t(),i(v(_.Component),{key:0,gateway:n.gateway},null,8,["gateway"])):(t(),i(s,{key:1,items:n.inbounds,predicate:p=>`${p.port}`===a.params.service.split(":")[1],find:!0},{default:o(({items:p})=>[(t(),i(v(_.Component),{inbound:p[0],gateway:n.gateway},null,8,["inbound","gateway"]))]),_:2},1032,["items","predicate"]))]),_:2},1024)]),_:2},1024)]),_:1})}}});export{S as default};
