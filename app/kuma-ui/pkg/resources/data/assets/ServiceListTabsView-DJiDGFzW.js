import{L as v}from"./LinkBox-DfSI8I6U.js";import{d as w,a as s,o as i,b as l,w as t,e as o,m as f,c as V,F as R,G as h,n as k,f as n,t as x}from"./index-B_EoIyfE.js";const g=w({__name:"ServiceListTabsView",setup(L){return(b,B)=>{const m=s("RouteTitle"),u=s("RouterLink"),p=s("RouterView"),_=s("AppView"),d=s("RouteView");return i(),l(d,{name:"service-list-tabs-view",params:{mesh:""}},{default:t(({route:a,t:r})=>[o(_,null,{title:t(()=>{var e;return[f("h2",null,[o(m,{title:r(`${((e=a.active)==null?void 0:e.name)==="service-list-view"?"":"external-"}services.routes.items.title`)},null,8,["title"])])]}),actions:t(()=>[o(v,null,{default:t(()=>[(i(!0),V(R,null,h(a.children,({name:e})=>{var c;return i(),l(u,{key:e,class:k({active:((c=a.active)==null?void 0:c.name)===e}),to:{name:e,params:{mesh:a.params.mesh}},"data-testid":`${e}-sub-tab`},{default:t(()=>[n(x(r(`services.routes.items.navigation.${e}`)),1)]),_:2},1032,["class","to","data-testid"])}),128))]),_:2},1024)]),default:t(()=>[n(),n(),o(p)]),_:2},1024)]),_:1})}}});export{g as default};
