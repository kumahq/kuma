import{d as A,r as t,o as a,p as i,w as e,b as o,e as c,m as w,P as V,l as E,c as p,J as _,K as b,Q as N,t as d,q as k}from"./index-gI7YoWPY.js";import{_ as R}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-CfFLyldn.js";const F=A({__name:"HostnameGeneratorDetailView",setup($){return(j,r)=>{const f=t("RouteTitle"),B=t("XCopyButton"),L=t("XBadge"),h=t("XLayout"),X=t("XAboutCard"),x=t("DataSource"),D=t("AppView"),v=t("DataLoader"),S=t("RouteView");return a(),i(S,{name:"hostname-generator-detail-view",params:{name:""}},{default:e(({route:g,t:l,uri:y})=>[o(f,{title:l("hostname-generators.routes.items.title"),render:!1},null,8,["title"]),r[3]||(r[3]=c()),o(v,{src:y(w(V),"/hostname-generators/:name",{name:g.params.name})},{default:e(({data:n})=>[o(D,{docs:l("hostname-generators.href.docs")},{title:e(()=>[E("h1",null,[o(B,{text:n.name},{default:e(()=>[o(f,{title:l("hostname-generators.routes.item.title",{name:n.name})},null,8,["title"])]),_:2},1032,["text"])])]),default:e(()=>[r[2]||(r[2]=c()),o(h,{type:"stack"},{default:e(()=>[o(X,{title:l("hostname-generators.routes.item.about.title"),created:n.creationTime,modified:n.modificationTime},{default:e(()=>[(a(!0),p(_,null,b([{...n.spec.selector.meshService.matchLabels,...n.spec.selector.meshExternalService.matchLabels,...n.spec.selector.meshMultiZoneService.matchLabels}],s=>(a(),p(_,{key:typeof s},[Object.keys(s).length?(a(),i(N,{key:0,layout:"horizontal"},{title:e(()=>[c(d(l("http.api.property.tags")),1)]),body:e(()=>[o(h,{type:"separated"},{default:e(()=>[(a(!0),p(_,null,b(s,(u,m)=>(a(),i(L,{key:m},{default:e(()=>[c(d(m)+":"+d(u),1)]),_:2},1024))),128))]),_:2},1024)]),_:2},1024)):k("",!0)],64))),128))]),_:2},1032,["title","created","modified"]),r[1]||(r[1]=c()),o(R,{resource:n.$raw},{default:e(({copy:s,copying:u})=>[u?(a(),i(x,{key:0,src:y(w(V),"/hostname-generators/:name/as/kubernetes",{name:g.params.name},{cacheControl:"no-store"}),onChange:m=>{s(C=>C(m))},onError:m=>{s((C,T)=>T(m))}},null,8,["src","onChange","onError"])):k("",!0)]),_:2},1032,["resource"])]),_:2},1024)]),_:2},1032,["docs"])]),_:2},1032,["src"])]),_:1})}}});export{F as default};
