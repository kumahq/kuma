import{d as R,e as t,o as r,m as l,w as e,a as n,k as S,t as p,b as c,R as v,J as y,E as A,r as D}from"./index-bM6gVJZj.js";const g=R({__name:"SubscriptionSummaryView",props:{data:{}},setup(d){const m=d;return(u,I)=>{const _=t("XAction"),f=t("XTabs"),b=t("RouterView"),w=t("AppView"),V=t("DataCollection"),C=t("RouteView");return r(),l(C,{name:"subscription-summary-view",params:{subscription:""}},{default:e(({route:s,t:h})=>[n(V,{items:m.data,predicate:o=>o.id===s.params.subscription},{item:e(({item:o})=>[n(w,null,{title:e(()=>[S("h2",null,p(o.zoneInstanceId??o.globalInstanceId),1)]),default:e(()=>{var i;return[c(),n(f,{selected:(i=s.child())==null?void 0:i.name},v({_:2},[y(s.children,({name:a})=>({name:`${a}-tab`,fn:e(()=>[n(_,{to:{name:a}},{default:e(()=>[c(p(h(`subscriptions.routes.item.navigation.${a}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),c(),n(b,null,{default:e(({Component:a})=>[(r(),l(A(a),{data:o},{default:e(()=>[D(u.$slots,"default")]),_:2},1032,["data"]))]),_:2},1024)]}),_:2},1024)]),_:2},1032,["items","predicate"])]),_:3})}}});export{g as default};
