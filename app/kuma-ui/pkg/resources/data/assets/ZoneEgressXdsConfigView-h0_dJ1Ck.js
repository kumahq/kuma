import{d as h,r as n,o as E,q as k,w as t,b as s,e as d,p as R,ar as w}from"./index-oTPgN0we.js";const z=h({__name:"ZoneEgressXdsConfigView",setup(X){return(y,a)=>{const r=n("RouteTitle"),c=n("XCheckbox"),i=n("XAction"),l=n("XCodeBlock"),p=n("DataLoader"),m=n("XCard"),g=n("AppView"),u=n("RouteView");return E(),k(u,{name:"zone-egress-xds-config-view",params:{zoneEgress:"",codeSearch:"",codeFilter:!1,codeRegExp:!1,includeEds:!1}},{default:t(({route:e,t:_,uri:f})=>[s(r,{render:!1,title:_("zone-egresses.routes.item.navigation.zone-egress-xds-config-view")},null,8,["title"]),a[2]||(a[2]=d()),s(g,null,{default:t(()=>[s(m,null,{default:t(()=>[s(p,{src:f(R(w),"/zone-egresses/:name/xds/:endpoints",{name:e.params.zoneEgress,endpoints:String(e.params.includeEds)})},{default:t(({data:C,refresh:x})=>[s(l,{language:"json",code:JSON.stringify(C,null,2),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},{"primary-actions":t(()=>[s(c,{checked:e.params.includeEds,label:"Include Endpoints",onChange:o=>e.update({includeEds:o})},null,8,["checked","onChange"]),a[1]||(a[1]=d()),s(i,{action:"refresh",appearance:"primary",onClick:x},{default:t(()=>a[0]||(a[0]=[d(`
                Refresh
              `)])),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{z as default};
