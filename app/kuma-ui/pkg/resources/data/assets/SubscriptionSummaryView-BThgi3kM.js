import{d as h,e as a,o as l,k as p,w as e,a as n,i as N,t as d,b as c,Q as S,G as A,C as D,r as R}from"./index-loxRIpgb.js";const k=h({__name:"SubscriptionSummaryView",props:{data:{},routeName:{}},setup(u){const r=u;return(_,X)=>{const m=a("XAction"),f=a("XTabs"),V=a("RouterView"),b=a("AppView"),w=a("DataCollection"),C=a("RouteView");return l(),p(C,{name:r.routeName,params:{subscription:""}},{default:e(({route:s,t:I})=>[n(w,{items:r.data,predicate:t=>t.id===s.params.subscription},{item:e(({item:t})=>[n(b,null,{title:e(()=>[N("h2",null,d(t.zoneInstanceId??t.globalInstanceId??t.controlPlaneInstanceId),1)]),default:e(()=>{var i;return[c(),n(f,{selected:(i=s.child())==null?void 0:i.name},S({_:2},[A(s.children,({name:o})=>({name:`${o}-tab`,fn:e(()=>[n(m,{to:{name:o}},{default:e(()=>[c(d(I(`subscriptions.routes.item.navigation.${o}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),c(),n(V,null,{default:e(({Component:o})=>[(l(),p(D(o),{data:t},{default:e(()=>[R(_.$slots,"default")]),_:2},1032,["data"]))]),_:2},1024)]}),_:2},1024)]),_:2},1032,["items","predicate"])]),_:3},8,["name"])}}});export{k as default};