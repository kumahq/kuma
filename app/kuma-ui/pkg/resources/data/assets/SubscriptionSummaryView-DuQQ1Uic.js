import{d as S,e as o,o as p,m as d,w as e,a,k as A,t as m,b as i,Q as D,J as R,E as X,r as g}from"./index-C_eW3RRu.js";const v=S({__name:"SubscriptionSummaryView",props:{data:{},routeName:{}},setup(u){const c=u;return(_,s)=>{const f=o("XAction"),V=o("XTabs"),b=o("RouterView"),w=o("AppView"),C=o("DataCollection"),I=o("RouteView");return p(),d(I,{name:c.routeName,params:{subscription:""}},{default:e(({route:r,t:N})=>[a(C,{items:c.data,predicate:t=>t.id===r.params.subscription},{item:e(({item:t})=>[a(w,null,{title:e(()=>[A("h2",null,m(t.zoneInstanceId??t.globalInstanceId??t.controlPlaneInstanceId),1)]),default:e(()=>{var l;return[s[0]||(s[0]=i()),a(V,{selected:(l=r.child())==null?void 0:l.name},D({_:2},[R(r.children,({name:n})=>({name:`${n}-tab`,fn:e(()=>[a(f,{to:{name:n}},{default:e(()=>[i(m(N(`subscriptions.routes.item.navigation.${n}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),s[1]||(s[1]=i()),a(b,null,{default:e(({Component:n})=>[(p(),d(X(n),{data:t},{default:e(()=>[g(_.$slots,"default")]),_:2},1032,["data"]))]),_:2},1024)]}),_:2},1024)]),_:2},1032,["items","predicate"])]),_:3},8,["name"])}}});export{v as default};
